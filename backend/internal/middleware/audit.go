package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/ipdb"
)

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	auditRepo domain.AuditRepo
	ipdb      *ipdb.IPDB
	logger    *slog.Logger
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(
	auditRepo domain.AuditRepo,
	ipdb *ipdb.IPDB,
	logger *slog.Logger,
) *AuditMiddleware {
	return &AuditMiddleware{
		auditRepo: auditRepo,
		ipdb:      ipdb,
		logger:    logger.With("middleware", "audit"),
	}
}

// responseWriter 包装响应写入器以捕获响应内容
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
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

			// 读取请求体
			var requestBody []byte
			if c.Request().Body != nil {
				requestBody, _ = io.ReadAll(c.Request().Body)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			// 包装响应写入器
			responseBody := &bytes.Buffer{}
			wrapper := &responseWriter{
				ResponseWriter: c.Response().Writer,
				body:           responseBody,
				statusCode:     200,
			}
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
			m.recordAuditLog(
				c.Request().Context(),
				c,
				requestBody,
				responseBody.Bytes(),
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

// recordAuditLog 记录审计日志
func (m *AuditMiddleware) recordAuditLog(ctx context.Context, c echo.Context, requestBody, responseBody []byte, statusCode int, duration time.Duration, handlerErr error) {
	// 获取基本信息
	requestID := m.getRequestID(c)
	sessionID := m.getSessionID(c)
	operatorType, operatorID, operatorName := m.getOperatorInfo(c)
	operationType := m.parseOperationType(c.Request().Method, c.Request().URL.Path)
	resourceType, resourceID, resourceName := m.parseResourceInfo(c)

	// 获取IP地理位置信息
	clientIP := m.getClientIP(c.Request())
	country, province, city := m.getLocationInfo(clientIP)

	// 处理请求体和响应体
	var requestBodyStr, responseBodyStr *string
	if len(requestBody) > 0 {
		sanitized := m.sanitizeData(string(requestBody))
		requestBodyStr = &sanitized
	}
	if len(responseBody) > 0 {
		sanitized := m.sanitizeData(string(responseBody))
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
		Country:        country,
		Province:       province,
		City:           city,
		Status:         status,
		ErrorMessage:   m.getErrorMessage(handlerErr),
		BusinessData:   make(map[string]interface{}),
		CreatedAt:      time.Now().Unix(),
	}

	// 异步记录审计日志
	go func() {
		// 这里需要调用具体的仓库实现，暂时跳过
		m.logger.Info("审计日志记录",
			"request_id", requestID,
			"operator_type", operatorType,
			"operation_type", operationType,
			"resource_type", resourceType,
			"status", status,
			"audit_log", auditLog,
		)
	}()
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
	// 从上下文中获取用户信息
	if user := c.Get("user"); user != nil {
		// 假设用户对象有ID和Name字段
		if userMap, ok := user.(map[string]interface{}); ok {
			if id, exists := userMap["id"]; exists {
				idStr := fmt.Sprintf("%v", id)
				name := ""
				if n, exists := userMap["name"]; exists {
					name = fmt.Sprintf("%v", n)
				}
				return consts.OperatorTypeUser, &idStr, &name
			}
		}
	}

	// 检查是否是管理员
	if admin := c.Get("admin"); admin != nil {
		if adminMap, ok := admin.(map[string]interface{}); ok {
			if id, exists := adminMap["id"]; exists {
				idStr := fmt.Sprintf("%v", id)
				name := ""
				if n, exists := adminMap["name"]; exists {
					name = fmt.Sprintf("%v", n)
				}
				return consts.OperatorTypeAdmin, &idStr, &name
			}
		}
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

	for _, field := range sensitiveFields {
		if strings.Contains(strings.ToLower(data), field) {
			return "[SENSITIVE_DATA_REMOVED]"
		}
	}

	// 限制数据长度
	if len(data) > 1000 {
		return data[:1000] + "..."
	}

	return data
}

// getErrorMessage 获取错误信息
func (m *AuditMiddleware) getErrorMessage(err error) *string {
	if err != nil {
		msg := err.Error()
		return &msg
	}
	return nil
}
