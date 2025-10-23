package credential

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// NewCredentialVault 创建凭证保险库实例
func NewCredentialVault(cfg *config.Config) (domain.CredentialVault, error) {
	if cfg.Credential.EncryptionKey == "" {
		return nil, fmt.Errorf("credential encryption key is required")
	}

	if len(cfg.Credential.EncryptionKey) != 32 {
		return nil, fmt.Errorf("credential encryption key must be exactly 32 bytes for AES-256, got %d bytes", len(cfg.Credential.EncryptionKey))
	}

	return NewVault(cfg.Credential.EncryptionKey)
}

// Vault 凭证加密存储实现
type Vault struct {
	key []byte // AES-256 密钥 (32字节)
}

// NewVault 创建新的凭证保险库
func NewVault(key string) (*Vault, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be exactly 32 bytes for AES-256")
	}

	return &Vault{
		key: []byte(key),
	}, nil
}

// Encrypt 加密凭证信息
func (v *Vault) Encrypt(ctx context.Context, credential map[string]interface{}) (map[string]interface{}, error) {
	if credential == nil {
		return nil, fmt.Errorf("credential cannot be nil")
	}

	// 将凭证序列化为JSON
	data, err := json.Marshal(credential)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// 返回base64编码的加密数据
	encrypted := base64.StdEncoding.EncodeToString(ciphertext)

	return map[string]interface{}{
		"encrypted_data": encrypted,
		"algorithm":      "AES-256-GCM",
	}, nil
}

// Decrypt 解密凭证信息
func (v *Vault) Decrypt(ctx context.Context, encryptedCredential map[string]interface{}) (map[string]interface{}, error) {
	if encryptedCredential == nil {
		return nil, fmt.Errorf("encrypted credential cannot be nil")
	}

	// 获取加密数据
	encryptedDataStr, ok := encryptedCredential["encrypted_data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid encrypted data format")
	}

	// 验证算法
	algorithm, ok := encryptedCredential["algorithm"].(string)
	if !ok || algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedDataStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 检查密文长度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// 提取nonce和密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// 反序列化JSON
	var credential map[string]interface{}
	if err := json.Unmarshal(plaintext, &credential); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
	}

	return credential, nil
}
