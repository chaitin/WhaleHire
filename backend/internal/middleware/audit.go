package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/ipdb"
	"github.com/chaitin/WhaleHire/backend/pkg/session"
)

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	auditRepo domain.AuditRepo
	ipdb      *ipdb.IPDB
	logger    *slog.Logger
	session   *session.Session
}

const (
	// maxAuditBodyBytes 限制审计记录保存的请求和响应体大小，避免大对象导致内存和延迟问题
	maxAuditBodyBytes = 4 * 1024
	// auditLogWriteTimeout 限制审计写库超时时间，防止后台协程长时间阻塞
	auditLogWriteTimeout = 5 * time.Second
)

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(
	auditRepo domain.AuditRepo,
	ipdb *ipdb.IPDB,
	logger *slog.Logger,
	session *session.Session,
) *AuditMiddleware {
	return &AuditMiddleware{
		auditRepo: auditRepo,
		ipdb:      ipdb,
		logger:    logger.With("middleware", "audit"),
		session:   session,
	}
}

// responseWriter 包装响应写入器以捕获响应内容
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	maxBytes   int
	statusCode int
	truncated  bool
	flusher    http.Flusher
	hijacker   http.Hijacker
	pusher     http.Pusher
}

func newResponseWriter(w http.ResponseWriter, maxBytes int, captureBody bool) *responseWriter {
	var bodyBuf *bytes.Buffer
	if captureBody {
		bodyBuf = &bytes.Buffer{}
	}

	rw := &responseWriter{
		ResponseWriter: w,
		body:           bodyBuf,
		maxBytes:       maxBytes,
		statusCode:     http.StatusOK,
	}

	if f, ok := w.(http.Flusher); ok {
		rw.flusher = f
	}
	if h, ok := w.(http.Hijacker); ok {
		rw.hijacker = h
	}
	if p, ok := w.(http.Pusher); ok {
		rw.pusher = p
	}

	return rw
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.body != nil && w.maxBytes > 0 && !w.truncated {
		remaining := w.maxBytes - w.body.Len()
		if remaining > 0 {
			if len(b) > remaining {
				w.body.Write(b[:remaining])
				w.truncated = true
			} else {
				w.body.Write(b)
			}
		} else {
			w.truncated = true
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Flush() {
	if w.flusher != nil {
		w.flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.hijacker == nil {
		return nil, nil, errors.New("底层响应写入器不支持 Hijack")
	}
	return w.hijacker.Hijack()
}

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if w.pusher == nil {
		return http.ErrNotSupported
	}
	return w.pusher.Push(target, opts)
}

func (w *responseWriter) capturedBody() ([]byte, bool) {
	if w.body == nil {
		return nil, false
	}
	return w.body.Bytes(), w.truncated
}

// Audit 审计中间件函数
func (m *AuditMiddleware) Audit() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 跳过不需要审计的路径
			if m.shouldSkip(c.Request().URL.Path) {
				return next(c)
			}

			// 跳过 GET 请求（查看操作）
			if c.Request().Method == "GET" {
				return next(c)
			}

			// 采集请求体（限制长度，避免破坏流式处理）
			var (
				requestBody          []byte
				requestBodyTruncated bool
			)
			requestID := m.getRequestID(c)
			sessionID := m.getSessionID(c)
			operatorType, operatorID, operatorName := m.getOperatorInfo(c)

			if req := c.Request(); req.Body != nil && m.shouldCaptureRequestBody(req) {
				var err error
				requestBody, requestBodyTruncated, err = m.captureRequestBody(req)
				if err != nil {
					m.logger.Warn("读取请求体用于审计失败", "path", req.URL.Path, "error", err)
				}
			}

			// 包装响应写入器
			wrapper := newResponseWriter(c.Response().Writer, maxAuditBodyBytes, m.shouldCaptureResponseBody(c))
			c.Response().Writer = wrapper

			// 记录开始时间
			startTime := time.Now()

			// 执行处理器
			var handlerErr error
			if err := next(c); err != nil {
				handlerErr = err
			}

			// 计算耗时
			duration := time.Since(startTime)

			// 记录审计日志
			responseBody, responseTruncated := wrapper.capturedBody()
			m.recordAuditLog(
				c.Request().Context(),
				c,
				requestID,
				sessionID,
				operatorType,
				operatorID,
				operatorName,
				requestBody,
				requestBodyTruncated,
				responseBody,
				responseTruncated,
				wrapper.statusCode,
				duration,
				handlerErr,
			)

			return handlerErr
		}
	}
}

// shouldSkip 判断是否跳过审计
func (m *AuditMiddleware) shouldSkip(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/api/v1/ping",
		"/api/v1/health",
		"/swagger",
		"/docs",
		"/static",
		"/assets",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldCaptureRequestBody 判断是否需要采集请求体
func (m *AuditMiddleware) shouldCaptureRequestBody(req *http.Request) bool {
	contentType := req.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") ||
		strings.HasPrefix(contentType, "application/octet-stream") {
		return false
	}
	// 允许业务显式跳过审计（例如大文件上传）
	if strings.EqualFold(req.Header.Get("X-Audit-Skip-Body"), "true") {
		return false
	}
	return true
}

// shouldCaptureResponseBody 判断是否需要采集响应体
func (m *AuditMiddleware) shouldCaptureResponseBody(c echo.Context) bool {
	req := c.Request()
	if req == nil {
		return true
	}
	if strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}
	path := req.URL.Path
	if strings.HasPrefix(path, "/api/v1/file") || strings.HasPrefix(path, "/api/v1/download") {
		return false
	}
	return true
}

type replayReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (r *replayReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *replayReadCloser) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

// captureRequestBody 截断采集请求体，同时保障业务仍可读取完整内容
func (m *AuditMiddleware) captureRequestBody(req *http.Request) ([]byte, bool, error) {
	original := req.Body
	if original == nil {
		return nil, false, nil
	}

	limitedReader := io.LimitReader(original, maxAuditBodyBytes+1)
	captured, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, false, err
	}

	truncated := len(captured) > maxAuditBodyBytes
	logged := captured
	if truncated {
		logged = captured[:maxAuditBodyBytes]
	}

	req.Body = &replayReadCloser{
		reader: io.MultiReader(bytes.NewReader(captured), original),
		closer: original,
	}

	return logged, truncated, nil
}

// recordAuditLog 记录审计日志
func (m *AuditMiddleware) recordAuditLog(
	ctx context.Context,
	c echo.Context,
	requestID string,
	sessionID string,
	operatorType consts.OperatorType,
	operatorID *string,
	operatorName *string,
	requestBody []byte,
	requestTruncated bool,
	responseBody []byte,
	responseTruncated bool,
	statusCode int,
	duration time.Duration,
	handlerErr error,
) {
	operationType := m.parseOperationType(c.Request().Method, c.Request().URL.Path)
	resourceType, resourceID, resourceName := m.parseResourceInfo(c)

	// 获取客户端IP
	clientIP := m.getClientIP(c.Request())

	// 处理请求体和响应体
	var requestBodyStr, responseBodyStr *string
	if len(requestBody) > 0 {
		sanitized := m.sanitizeData(string(requestBody))
		if requestTruncated {
			sanitized += "（内容已截断）"
		}
		requestBodyStr = &sanitized
	}
	if len(responseBody) > 0 {
		sanitized := m.sanitizeData(string(responseBody))
		if responseTruncated {
			sanitized += "（内容已截断）"
		}
		responseBodyStr = &sanitized
	}

	// 处理查询参数
	var queryStr *string
	if c.Request().URL.RawQuery != "" {
		queryStr = &c.Request().URL.RawQuery
	}

	// 处理用户代理
	userAgent := c.Request().UserAgent()
	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// 确定状态
	var status consts.AuditLogStatus
	if handlerErr != nil || statusCode >= 400 {
		status = consts.AuditLogStatusFailed
	} else {
		status = consts.AuditLogStatusSuccess
	}

	// 构建审计日志
	auditLog := &domain.AuditLog{
		RequestID:      &requestID,
		SessionID:      &sessionID,
		OperatorType:   operatorType,
		OperatorID:     operatorID,
		OperatorName:   operatorName,
		OperationType:  operationType,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		ResourceName:   resourceName,
		RequestMethod:  c.Request().Method,
		RequestPath:    c.Request().URL.Path,
		RequestQuery:   queryStr,
		RequestBody:    requestBodyStr,
		ResponseStatus: statusCode,
		ResponseBody:   responseBodyStr,
		Duration:       int64(duration.Milliseconds()),
		IP:             clientIP,
		UserAgent:      userAgentPtr,
		Status:         status,
		ErrorMessage:   m.getErrorMessage(handlerErr),
		BusinessData:   make(map[string]interface{}),
		CreatedAt:      time.Now().Unix(),
	}

	// 异步记录审计日志
	go func(parent context.Context, log *domain.AuditLog, ip string) {
		ctxDB := context.WithoutCancel(parent)
		ctxDB, cancel := context.WithTimeout(ctxDB, auditLogWriteTimeout)
		defer cancel()

		if country, province, city := m.getLocationInfo(ip); country != nil || province != nil || city != nil {
			log.Country = country
			log.Province = province
			log.City = city
		}

		if err := m.auditRepo.Create(ctxDB, log); err != nil {
			m.logger.Error("审计日志记录失败",
				"request_id", requestID,
				"operator_type", operatorType,
				"operation_type", operationType,
				"resource_type", resourceType,
				"status", status,
				"error", err,
			)
		}
	}(ctx, auditLog, clientIP)
}

// getRequestID 获取请求ID
func (m *AuditMiddleware) getRequestID(c echo.Context) string {
	if requestID := c.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		return requestID
	}
	return uuid.New().String()
}

// getSessionID 获取会话ID
func (m *AuditMiddleware) getSessionID(c echo.Context) string {

	// 尝试从 cookie 中获取会话ID
	if cookie, err := c.Cookie(consts.SessionName); err == nil {
		return cookie.Value
	}
	if cookie, err := c.Cookie(consts.UserSessionName); err == nil {
		return cookie.Value
	}

	// 如果都没有，返回空字符串
	return ""
}

// getOperatorInfo 获取操作者信息
func (m *AuditMiddleware) getOperatorInfo(c echo.Context) (consts.OperatorType, *string, *string) {
	// 首先尝试获取管理员信息
	if admin, err := session.Get[domain.AdminUser](m.session, c, consts.SessionName); err == nil {
		return consts.OperatorTypeAdmin, &admin.ID, &admin.Username
	}

	// 然后尝试获取普通用户信息
	if user, err := session.Get[domain.User](m.session, c, consts.UserSessionName); err == nil {
		return consts.OperatorTypeUser, &user.ID, &user.Username
	}

	// 默认为用户类型，无ID和名称
	return consts.OperatorTypeUser, nil, nil
}

// parseOperationType 解析操作类型
func (m *AuditMiddleware) parseOperationType(method, path string) consts.OperationType {
	if method == http.MethodPost {
		switch {
		case strings.HasSuffix(path, "/login"):
			return consts.OperationTypeLogin
		case strings.HasSuffix(path, "/logout"):
			return consts.OperationTypeLogout
		}
	}

	switch method {
	case http.MethodGet:
		return consts.OperationTypeView
	case http.MethodPost:
		return consts.OperationTypeCreate
	case http.MethodPut, http.MethodPatch:
		return consts.OperationTypeUpdate
	case http.MethodDelete:
		return consts.OperationTypeDelete
	default:
		return consts.OperationTypeView
	}
}

// parseResourceInfo 解析资源信息
func (m *AuditMiddleware) parseResourceInfo(c echo.Context) (consts.ResourceType, *string, *string) {
	path := c.Request().URL.Path

	switch {
	case strings.HasPrefix(path, "/api/v1/admin/setting"):
		return consts.ResourceTypeSetting, nil, nil
	case strings.HasPrefix(path, "/api/v1/admin/role"):
		return consts.ResourceTypeRole, m.extractIDFromPath(path, "role"), nil
	case strings.HasPrefix(path, "/api/v1/admin"):
		return m.parseAdminResource(c, path)
	case strings.HasPrefix(path, "/api/v1/user"):
		return m.parseUserResource(c, path)
	case strings.HasPrefix(path, "/api/v1/users"):
		return m.parseUserResource(c, path)
	case strings.HasPrefix(path, "/api/v1/resume"):
		return m.parseResumeResource(c, path)
	case strings.HasPrefix(path, "/api/v1/resumes"):
		return m.parseResumeResource(c, path)
	case strings.HasPrefix(path, "/api/v1/departments"):
		return m.parseDepartmentResource(c, path)
	case strings.HasPrefix(path, "/api/v1/job-profiles") || strings.HasPrefix(path, "/api/v1/job-skills"):
		return m.parseJobProfileResource(c, path)
	case strings.HasPrefix(path, "/api/v1/job-applications"):
		return m.parseJobApplicationResource(c, path)
	case strings.HasPrefix(path, "/api/v1/file"):
		return m.parseFileResource(c, path)
	case strings.HasPrefix(path, "/api/v1/screening"):
		return m.parseScreeningResource(c, path)
	case strings.HasPrefix(path, "/api/v1/general-agent/conversations"):
		return m.parseGeneralAgentConversationResource(c, path)
	case strings.HasPrefix(path, "/api/v1/general-agent"):
		return m.parseGeneralAgentResource(c, path)
	default:
		return consts.ResourceTypeUser, nil, nil
	}
}

// parseUserResource 解析用户资源
func (m *AuditMiddleware) parseUserResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if userID := c.Param("id"); userID != "" {
		id := userID
		return consts.ResourceTypeUser, &id, nil
	}
	if userID := c.QueryParam("id"); userID != "" {
		id := userID
		return consts.ResourceTypeUser, &id, nil
	}
	return consts.ResourceTypeUser, nil, nil
}

// parseAdminResource 解析管理员资源
func (m *AuditMiddleware) parseAdminResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if adminID := c.Param("id"); adminID != "" {
		id := adminID
		return consts.ResourceTypeAdmin, &id, nil
	}
	if adminID := c.QueryParam("id"); adminID != "" {
		id := adminID
		return consts.ResourceTypeAdmin, &id, nil
	}
	return consts.ResourceTypeAdmin, m.extractIDFromPath(path, "admin"), nil
}

// parseResumeResource 解析简历资源
func (m *AuditMiddleware) parseResumeResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if resumeID := c.Param("id"); resumeID != "" {
		id := resumeID
		return consts.ResourceTypeResume, &id, nil
	}
	if resumeID := c.Param("resume_id"); resumeID != "" {
		id := resumeID
		return consts.ResourceTypeResume, &id, nil
	}
	return consts.ResourceTypeResume, m.extractIDFromPath(path, "resume", "resumes"), nil
}

// parseDepartmentResource 解析部门资源
func (m *AuditMiddleware) parseDepartmentResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	return consts.ResourceTypeDepartment, m.extractIDFromPath(path, "departments"), nil
}

// parseJobProfileResource 解析职位配置资源
func (m *AuditMiddleware) parseJobProfileResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if id := c.Param("id"); id != "" {
		idCopy := id
		return consts.ResourceTypeJobPosition, &idCopy, nil
	}
	return consts.ResourceTypeJobPosition, m.extractIDFromPath(path, "job-profiles", "meta"), nil
}

// parseJobApplicationResource 解析岗位申请资源
func (m *AuditMiddleware) parseJobApplicationResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if resumeID := c.Param("resume_id"); resumeID != "" {
		idCopy := resumeID
		return consts.ResourceTypeResume, &idCopy, nil
	}
	if jobPositionID := c.Param("job_position_id"); jobPositionID != "" {
		idCopy := jobPositionID
		return consts.ResourceTypeJobPosition, &idCopy, nil
	}
	return consts.ResourceTypeJobPosition, nil, nil
}

// parseFileResource 解析附件资源
func (m *AuditMiddleware) parseFileResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if key := c.Param("key"); key != "" {
		keyCopy := key
		return consts.ResourceTypeAttachment, &keyCopy, nil
	}
	return consts.ResourceTypeAttachment, m.extractIDFromPath(path, "stream"), nil
}

// parseScreeningResource 解析筛选任务资源
func (m *AuditMiddleware) parseScreeningResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if id := c.Param("id"); id != "" {
		idCopy := id
		return consts.ResourceTypeScreening, &idCopy, nil
	}
	if taskID := c.Param("task_id"); taskID != "" {
		idCopy := taskID
		return consts.ResourceTypeScreening, &idCopy, nil
	}
	return consts.ResourceTypeScreening, m.extractIDFromPath(path, "tasks"), nil
}

// parseGeneralAgentConversationResource 解析通用智能体对话资源
func (m *AuditMiddleware) parseGeneralAgentConversationResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	if strings.HasSuffix(path, "/addmessage") {
		if id := c.Param("id"); id != "" {
			idCopy := id
			return consts.ResourceTypeMessage, &idCopy, nil
		}
		return consts.ResourceTypeMessage, nil, nil
	}

	if id := c.Param("id"); id != "" {
		idCopy := id
		return consts.ResourceTypeConversation, &idCopy, nil
	}
	if id := c.QueryParam("id"); id != "" {
		idCopy := id
		return consts.ResourceTypeConversation, &idCopy, nil
	}
	return consts.ResourceTypeConversation, nil, nil
}

// parseGeneralAgentResource 解析通用智能体消息资源
func (m *AuditMiddleware) parseGeneralAgentResource(c echo.Context, path string) (consts.ResourceType, *string, *string) {
	return consts.ResourceTypeMessage, nil, nil
}

// extractIDFromPath 从路径中提取ID
func (m *AuditMiddleware) extractIDFromPath(path string, resources ...string) *string {
	if len(resources) == 0 {
		return nil
	}
	parts := strings.Split(path, "/")
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		for _, resource := range resources {
			if part == resource && i+1 < len(parts) {
				id := parts[i+1]
				if id == "" || id == resource {
					continue
				}
				idCopy := id
				return &idCopy
			}
		}
	}
	return nil
}

// getClientIP 获取客户端IP
func (m *AuditMiddleware) getClientIP(req *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查 X-Real-IP 头
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用 RemoteAddr
	ip := req.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

// getLocationInfo 获取IP地理位置信息
func (m *AuditMiddleware) getLocationInfo(ip string) (*string, *string, *string) {
	if m.ipdb == nil {
		return nil, nil, nil
	}

	ipInfo, err := m.ipdb.Lookup(ip)
	if err != nil {
		m.logger.Warn("获取IP地理位置失败", "ip", ip, "error", err)
		return nil, nil, nil
	}

	return &ipInfo.Country, &ipInfo.Province, &ipInfo.City
}

// sanitizeData 清理敏感数据
func (m *AuditMiddleware) sanitizeData(data string) string {
	// 简单的敏感数据清理
	sensitiveFields := []string{"password", "token", "secret", "key"}
	lower := strings.ToLower(data)

	for _, field := range sensitiveFields {
		if strings.Contains(lower, field) {
			return "[SENSITIVE_DATA_REMOVED]"
		}
	}

	// 限制数据长度，确保UTF-8安全
	if len(data) > 1000 {
		return m.truncateUTF8Safe(data, 1000) + "..."
	}

	return data
}

// truncateUTF8Safe 安全地截断UTF-8字符串，避免破坏多字节字符
func (m *AuditMiddleware) truncateUTF8Safe(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// 从maxBytes位置向前查找有效的UTF-8字符边界
	for i := maxBytes; i >= 0; i-- {
		if utf8.ValidString(s[:i]) {
			return s[:i]
		}
	}

	// 如果找不到有效边界，返回空字符串
	return ""
}

// getErrorMessage 获取错误信息
func (m *AuditMiddleware) getErrorMessage(err error) *string {
	if err != nil {
		msg := err.Error()
		return &msg
	}
	return nil
}
