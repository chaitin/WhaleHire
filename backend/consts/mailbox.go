package consts

// MailboxProtocol 邮箱协议类型
type MailboxProtocol string

const (
	MailboxProtocolIMAP MailboxProtocol = "imap" // IMAP协议
	MailboxProtocolPOP3 MailboxProtocol = "pop3" // POP3协议
)

// Values 返回所有邮箱协议值
func (MailboxProtocol) Values() []MailboxProtocol {
	return []MailboxProtocol{
		MailboxProtocolIMAP,
		MailboxProtocolPOP3,
	}
}

// IsValid 检查邮箱协议是否有效
func (m MailboxProtocol) IsValid() bool {
	for _, v := range MailboxProtocol("").Values() {
		if m == v {
			return true
		}
	}
	return false
}

// MailboxAuthType 邮箱认证类型
type MailboxAuthType string

const (
	MailboxAuthTypePassword MailboxAuthType = "password" // 密码认证
	MailboxAuthTypeOAuth    MailboxAuthType = "oauth"    // OAuth认证
)

// Values 返回所有邮箱认证类型值
func (MailboxAuthType) Values() []MailboxAuthType {
	return []MailboxAuthType{
		MailboxAuthTypePassword,
		MailboxAuthTypeOAuth,
	}
}

// IsValid 检查邮箱认证类型是否有效
func (m MailboxAuthType) IsValid() bool {
	for _, v := range MailboxAuthType("").Values() {
		if m == v {
			return true
		}
	}
	return false
}

// MailboxStatus 邮箱设置状态
type MailboxStatus string

const (
	MailboxStatusEnabled  MailboxStatus = "enabled"  // 启用
	MailboxStatusDisabled MailboxStatus = "disabled" // 禁用
)

// Values 返回所有邮箱状态值
func (MailboxStatus) Values() []MailboxStatus {
	return []MailboxStatus{
		MailboxStatusEnabled,
		MailboxStatusDisabled,
	}
}

// IsValid 检查邮箱状态是否有效
func (m MailboxStatus) IsValid() bool {
	for _, v := range MailboxStatus("").Values() {
		if m == v {
			return true
		}
	}
	return false
}

// ResumeSourceType 简历来源类型
type ResumeSourceType string

const (
	ResumeSourceTypeEmail  ResumeSourceType = "email"  // 邮箱采集
	ResumeSourceTypeManual ResumeSourceType = "manual" // 手动上传
)

// Values 返回所有简历来源类型值
func (ResumeSourceType) Values() []ResumeSourceType {
	return []ResumeSourceType{
		ResumeSourceTypeEmail,
		ResumeSourceTypeManual,
	}
}

// IsValid 检查简历来源类型是否有效
func (r ResumeSourceType) IsValid() bool {
	for _, v := range ResumeSourceType("").Values() {
		if r == v {
			return true
		}
	}
	return false
}
