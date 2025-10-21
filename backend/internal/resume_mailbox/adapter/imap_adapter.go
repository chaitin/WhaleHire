package adapter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
)

const (
	defaultIMAPTimeout = 30 * time.Second
	defaultFetchLimit  = 50
	imapMaxUID         = ^uint32(0)
)

type imapCursor struct {
	UIDValidity uint32 `json:"uid_validity"`
	LastUID     uint32 `json:"last_uid"`
}

// IMAPAdapter IMAP协议适配器
type IMAPAdapter struct {
	logger      *slog.Logger
	dialTimeout time.Duration
}

// NewIMAPAdapter 创建IMAP协议适配器
func NewIMAPAdapter(logger *slog.Logger) domain.MailboxProtocolAdapter {
	return &IMAPAdapter{
		logger:      logger,
		dialTimeout: defaultIMAPTimeout,
	}
}

// GetProtocol 返回协议标识
func (a *IMAPAdapter) GetProtocol() string {
	return string(consts.MailboxProtocolIMAP)
}

// TestConnection 测试连接是否可用
func (a *IMAPAdapter) TestConnection(ctx context.Context, config *domain.MailboxConnectionConfig) error {
	client, _, err := a.openMailbox(ctx, config, true)
	if client != nil {
		defer func() {
			_ = client.Logout()
		}()
	}
	return err
}

// Fetch 拉取新邮件
func (a *IMAPAdapter) Fetch(ctx context.Context, config *domain.MailboxConnectionConfig, req *domain.MailboxFetchRequest) (*domain.MailboxFetchResult, error) {
	if req == nil {
		req = &domain.MailboxFetchRequest{}
	}
	if req.Limit <= 0 {
		req.Limit = defaultFetchLimit
	}

	c, mailboxStatus, err := a.openMailbox(ctx, config, true)
	if err != nil {
		return nil, err
	}
	defer func() {
		if errLogout := c.Logout(); errLogout != nil {
			a.logger.Warn("IMAP 退出会话失败", slog.String("error", errLogout.Error()))
		}
	}()

	firstSync := strings.TrimSpace(req.Cursor) == ""
	currentCursor := a.parseCursor(req.Cursor)
	if currentCursor.UIDValidity != 0 && currentCursor.UIDValidity != mailboxStatus.UidValidity {
		// UIDVALIDITY 已变化，说明邮箱重置或被清空，需要重新同步
		a.logger.Info("检测到UIDVALIDITY变化，重置IMAP游标",
			slog.String("mailbox", config.EmailAddress),
			slog.Uint64("old_uid_validity", uint64(currentCursor.UIDValidity)),
			slog.Uint64("new_uid_validity", uint64(mailboxStatus.UidValidity)),
		)
		currentCursor.LastUID = 0
		currentCursor.UIDValidity = mailboxStatus.UidValidity
	}

	if firstSync {
		currentCursor.UIDValidity = mailboxStatus.UidValidity
		if mailboxStatus.UidNext > 0 {
			currentCursor.LastUID = mailboxStatus.UidNext - 1
		} else {
			currentCursor.LastUID = 0
		}

		nextCursor, errBuild := a.buildCursorString(currentCursor.UIDValidity, currentCursor.LastUID)
		if errBuild != nil {
			return nil, errBuild
		}

		return &domain.MailboxFetchResult{
			Messages:      []*domain.MailboxEmail{},
			NextCursor:    nextCursor,
			LastMessageID: "",
		}, nil
	}

	// 搜索新的UID列表
	searchSet := new(imap.SeqSet)
	if currentCursor.LastUID == 0 {
		searchSet.AddRange(1, imapMaxUID)
	} else {
		searchSet.AddRange(currentCursor.LastUID+1, imapMaxUID)
	}

	criteria := imap.NewSearchCriteria()
	criteria.Uid = searchSet

	uids, err := c.UidSearch(criteria)
	if err != nil {
		return nil, fmt.Errorf("搜索IMAP邮件失败: %w", err)
	}

	if len(uids) == 0 {
		// 没有新邮件
		nextCursor, _ := a.buildCursorString(mailboxStatus.UidValidity, currentCursor.LastUID)
		return &domain.MailboxFetchResult{
			Messages:      []*domain.MailboxEmail{},
			NextCursor:    nextCursor,
			LastMessageID: "",
		}, nil
	}

	sort.Slice(uids, func(i, j int) bool { return uids[i] < uids[j] })
	if len(uids) > req.Limit {
		uids = uids[len(uids)-req.Limit:]
	}

	fetchSet := new(imap.SeqSet)
	for _, uid := range uids {
		fetchSet.AddNum(uid)
	}

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchUid,
		imap.FetchRFC822Size,
		section.FetchItem(),
	}

	messages := make(chan *imap.Message, len(uids))
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(fetchSet, items, messages)
	}()

	result := &domain.MailboxFetchResult{
		Messages: []*domain.MailboxEmail{},
	}

	seen := make(map[string]struct{})
	var maxUID uint32
	var lastMessageID string

loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg, ok := <-messages:
			if !ok {
				break loop
			}

			if msg == nil {
				continue
			}
			if msg.Uid > maxUID {
				maxUID = msg.Uid
			}

			body := msg.GetBody(section)
			if body == nil {
				// 没有正文数据，跳过
				continue
			}

			email := &domain.MailboxEmail{
				Attachments: []*domain.MailboxAttachment{},
				RawSize:     int64(msg.Size),
			}

			if msg.Envelope != nil {
				email.Subject = msg.Envelope.Subject
				email.MessageID = strings.TrimSpace(msg.Envelope.MessageId)
				if !msg.Envelope.Date.IsZero() {
					email.ReceivedAt = msg.Envelope.Date
				}
			}

			reader, errCreate := mail.CreateReader(body)
			if errCreate != nil {
				a.logger.Warn("解析IMAP邮件失败",
					slog.String("mailbox", config.EmailAddress),
					slog.Uint64("uid", uint64(msg.Uid)),
					slog.String("error", errCreate.Error()),
				)
				continue
			}

			for {
				part, errNext := reader.NextPart()
				if errNext == io.EOF {
					break
				}
				if errNext != nil {
					a.logger.Warn("读取邮件分段失败",
						slog.String("mailbox", config.EmailAddress),
						slog.Uint64("uid", uint64(msg.Uid)),
						slog.String("error", errNext.Error()),
					)
					break
				}

				switch header := part.Header.(type) {
				case *mail.AttachmentHeader:
					filename, _ := header.Filename()
					filename = strings.TrimSpace(filename)
					if filename == "" {
						continue
					}

					contentType, _, _ := header.ContentType()

					buf := &bytes.Buffer{}
					size, copyErr := io.Copy(buf, part.Body)
					if copyErr != nil {
						a.logger.Warn("读取附件内容失败",
							slog.String("mailbox", config.EmailAddress),
							slog.Uint64("uid", uint64(msg.Uid)),
							slog.String("filename", filename),
							slog.String("error", copyErr.Error()),
						)
						continue
					}

					data := buf.Bytes()
					hash := sha256.Sum256(data)
					hashStr := hex.EncodeToString(hash[:])

					messageKey := email.MessageID
					if messageKey == "" {
						messageKey = strconv.FormatUint(uint64(msg.Uid), 10)
					}
					dedupKey := fmt.Sprintf("%s:%s", messageKey, hashStr)
					if _, exists := seen[dedupKey]; exists {
						continue
					}
					seen[dedupKey] = struct{}{}

					email.Attachments = append(email.Attachments, &domain.MailboxAttachment{
						Filename:    filename,
						ContentType: contentType,
						Content:     data,
						Size:        size,
						Hash:        hashStr,
						Reader:      bytes.NewReader(data),
					})
				}
			}

			if len(email.Attachments) > 0 {
				if email.MessageID != "" {
					lastMessageID = email.MessageID
				} else {
					lastMessageID = strconv.FormatUint(uint64(msg.Uid), 10)
				}
				result.Messages = append(result.Messages, email)
			}
		}
	}

	if fetchErr := <-done; fetchErr != nil {
		return nil, fmt.Errorf("拉取IMAP邮件失败: %w", fetchErr)
	}

	if maxUID < currentCursor.LastUID {
		maxUID = currentCursor.LastUID
	}

	nextCursor, err := a.buildCursorString(mailboxStatus.UidValidity, maxUID)
	if err != nil {
		return nil, err
	}

	result.NextCursor = nextCursor
	result.LastMessageID = lastMessageID

	return result, nil
}

// sendIMAPID 发送 IMAP ID 命令，部分邮箱（如网易系）要求客户端先上报身份信息
func (a *IMAPAdapter) sendIMAPID(c *client.Client, config *domain.MailboxConnectionConfig) {
	if c == nil || config == nil {
		return
	}

	supported, err := c.Support("ID")
	if err != nil {
		a.logger.Warn("查询IMAP服务器能力失败，跳过发送ID命令",
			slog.String("mailbox", config.EmailAddress),
			slog.String("error", err.Error()),
		)
		return
	}
	if !supported {
		a.logger.Debug("IMAP服务器未声明ID能力，跳过发送ID命令",
			slog.String("mailbox", config.EmailAddress),
		)
		return
	}

	payload := a.buildIMAPIDPayload(config)
	if len(payload) == 0 {
		return
	}

	cmd := &imap.Command{
		Name:      "ID",
		Arguments: []interface{}{payload},
	}

	status, execErr := c.Execute(cmd, nil)
	if execErr != nil {
		a.logger.Warn("发送IMAP ID命令失败，继续尝试连接",
			slog.String("mailbox", config.EmailAddress),
			slog.String("error", execErr.Error()),
		)
		return
	}
	if status != nil {
		if statusErr := status.Err(); statusErr != nil {
			a.logger.Warn("发送IMAP ID命令失败，继续尝试连接",
				slog.String("mailbox", config.EmailAddress),
				slog.String("error", statusErr.Error()),
			)
			return
		}
	}

	a.logger.Debug("IMAP ID命令发送成功",
		slog.String("mailbox", config.EmailAddress),
	)
}

func (a *IMAPAdapter) buildIMAPIDPayload(config *domain.MailboxConnectionConfig) []interface{} {
	if config == nil {
		return nil
	}

	values := []string{
		"name", "WhaleHire",
		"version", "1.0.0",
		"vendor", "Chaitin",
	}

	if email := strings.TrimSpace(config.EmailAddress); email != "" {
		values = append(values, "contact", email)
	}

	if len(values)%2 != 0 {
		// 仅保留成对键值，避免不合法的IMAP ID报文
		values = values[:len(values)-1]
	}

	list := make([]interface{}, len(values))
	for i, v := range values {
		list[i] = v
	}

	return list
}

func (a *IMAPAdapter) openMailbox(ctx context.Context, config *domain.MailboxConnectionConfig, readOnly bool) (*client.Client, *imap.MailboxStatus, error) {
	if config == nil {
		return nil, nil, fmt.Errorf("缺少邮箱连接配置")
	}

	addr := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	var (
		c   *client.Client
		err error
	)

	dialTimeout := a.dialTimeout
	if dialTimeout <= 0 {
		dialTimeout = defaultIMAPTimeout
	}

	dialer := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: 30 * time.Second,
	}

	if config.UseSSL {
		tlsConfig := &tls.Config{
			ServerName: config.Host,
		}
		conn, dialErr := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if dialErr != nil {
			return nil, nil, fmt.Errorf("无法建立IMAP TLS连接: %w", dialErr)
		}
		c, err = client.New(conn)
	} else {
		netConn, dialErr := dialer.DialContext(ctx, "tcp", addr)
		if dialErr != nil {
			return nil, nil, fmt.Errorf("无法连接IMAP服务器: %w", dialErr)
		}
		c, err = client.New(netConn)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("创建IMAP客户端失败: %w", err)
	}

	c.Timeout = dialTimeout

	go func() {
		<-ctx.Done()
		_ = c.Terminate()
	}()

	if errAuth := a.authenticate(c, config); errAuth != nil {
		_ = c.Logout()
		return nil, nil, errAuth
	}

	a.sendIMAPID(c, config)

	folder := strings.TrimSpace(config.Folder)
	if folder == "" {
		folder = "INBOX"
	}

	status, err := c.Select(folder, readOnly)
	if err != nil {
		_ = c.Logout()
		return nil, nil, fmt.Errorf("无法选择IMAP文件夹 %s: %w", folder, err)
	}

	return c, status, nil
}

func (a *IMAPAdapter) authenticate(c *client.Client, config *domain.MailboxConnectionConfig) error {
	switch strings.ToLower(config.AuthType) {
	case string(consts.MailboxAuthTypePassword):
		username := ""
		if v, ok := config.Credential["username"].(string); ok && v != "" {
			username = v
		}
		if username == "" {
			username = config.EmailAddress
		}
		password, _ := config.Credential["password"].(string)
		if username == "" || password == "" {
			return fmt.Errorf("缺少用户名或密码")
		}
		if err := c.Login(username, password); err != nil {
			return fmt.Errorf("IMAP 登录失败: %w", err)
		}
		return nil
	case string(consts.MailboxAuthTypeOAuth):
		username := ""
		if v, ok := config.Credential["username"].(string); ok && v != "" {
			username = v
		}
		if username == "" {
			username = config.EmailAddress
		}
		token, _ := config.Credential["access_token"].(string)
		if username == "" || token == "" {
			return fmt.Errorf("缺少OAuth凭证")
		}
		auth := &xoauth2Client{
			username: username,
			token:    token,
		}
		if err := c.Authenticate(auth); err != nil {
			return fmt.Errorf("IMAP OAuth2认证失败: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("不支持的认证类型: %s", config.AuthType)
	}
}

func (a *IMAPAdapter) parseCursor(cursor string) imapCursor {
	if cursor == "" {
		return imapCursor{}
	}
	var value imapCursor
	if err := json.Unmarshal([]byte(cursor), &value); err != nil {
		a.logger.Warn("解析IMAP游标失败，使用默认游标",
			slog.String("error", err.Error()),
		)
		return imapCursor{}
	}
	return value
}

func (a *IMAPAdapter) buildCursorString(uidValidity uint32, lastUID uint32) (string, error) {
	data, err := json.Marshal(imapCursor{
		UIDValidity: uidValidity,
		LastUID:     lastUID,
	})
	if err != nil {
		return "", fmt.Errorf("构建游标数据失败: %w", err)
	}
	return string(data), nil
}

type xoauth2Client struct {
	username string
	token    string
}

func (x *xoauth2Client) Start() (mech string, ir []byte, err error) {
	mech = "XOAUTH2"
	ir = []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", x.username, x.token))
	return
}

func (x *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	if len(challenge) == 0 {
		return nil, nil
	}
	return nil, sasl.ErrUnexpectedServerChallenge
}
