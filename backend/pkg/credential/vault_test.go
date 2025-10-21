package credential

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVault(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			key:     "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "invalid short key",
			key:     "short",
			wantErr: true,
		},
		{
			name:    "invalid long key",
			key:     "123456789012345678901234567890123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault, err := NewVault(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vault)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vault)
			}
		})
	}
}

func TestVault_EncryptDecrypt(t *testing.T) {
	vault, err := NewVault("12345678901234567890123456789012")
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name       string
		credential map[string]interface{}
	}{
		{
			name: "simple password credential",
			credential: map[string]interface{}{
				"username": "test@example.com",
				"password": "secret123",
			},
		},
		{
			name: "oauth credential",
			credential: map[string]interface{}{
				"access_token":  "oauth_token_123",
				"refresh_token": "refresh_token_456",
				"expires_in":    float64(3600),
			},
		},
		{
			name: "complex credential",
			credential: map[string]interface{}{
				"username": "user@domain.com",
				"password": "complex_password_!@#$%^&*()",
				"host":     "imap.gmail.com",
				"port":     float64(993),
				"ssl":      true,
				"metadata": map[string]interface{}{
					"created_at": "2024-01-01T00:00:00Z",
					"version":    "1.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := vault.Encrypt(ctx, tt.credential)
			require.NoError(t, err)
			assert.NotNil(t, encrypted)

			// 验证加密结果结构
			assert.Contains(t, encrypted, "encrypted_data")
			assert.Contains(t, encrypted, "algorithm")
			assert.Equal(t, "AES-256-GCM", encrypted["algorithm"])

			// 解密
			decrypted, err := vault.Decrypt(ctx, encrypted)
			require.NoError(t, err)
			assert.NotNil(t, decrypted)

			// 验证解密结果与原始数据一致
			assert.Equal(t, tt.credential, decrypted)
		})
	}
}

func TestVault_EncryptNilCredential(t *testing.T) {
	vault, err := NewVault("12345678901234567890123456789012")
	require.NoError(t, err)

	ctx := context.Background()

	encrypted, err := vault.Encrypt(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, encrypted)
	assert.Contains(t, err.Error(), "credential cannot be nil")
}

func TestVault_DecryptInvalidData(t *testing.T) {
	vault, err := NewVault("12345678901234567890123456789012")
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		encrypted map[string]interface{}
		wantErr   string
	}{
		{
			name:      "nil encrypted credential",
			encrypted: nil,
			wantErr:   "encrypted credential cannot be nil",
		},
		{
			name: "missing encrypted_data",
			encrypted: map[string]interface{}{
				"algorithm": "AES-256-GCM",
			},
			wantErr: "invalid encrypted data format",
		},
		{
			name: "missing algorithm",
			encrypted: map[string]interface{}{
				"encrypted_data": "some_data",
			},
			wantErr: "unsupported encryption algorithm",
		},
		{
			name: "unsupported algorithm",
			encrypted: map[string]interface{}{
				"encrypted_data": "some_data",
				"algorithm":      "AES-128-CBC",
			},
			wantErr: "unsupported encryption algorithm",
		},
		{
			name: "invalid base64",
			encrypted: map[string]interface{}{
				"encrypted_data": "invalid_base64!@#",
				"algorithm":      "AES-256-GCM",
			},
			wantErr: "failed to decode base64",
		},
		{
			name: "too short ciphertext",
			encrypted: map[string]interface{}{
				"encrypted_data": "dGVzdA==", // "test" in base64, too short for GCM
				"algorithm":      "AES-256-GCM",
			},
			wantErr: "ciphertext too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := vault.Decrypt(ctx, tt.encrypted)
			assert.Error(t, err)
			assert.Nil(t, decrypted)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestVault_EncryptDecryptConsistency(t *testing.T) {
	vault, err := NewVault("12345678901234567890123456789012")
	require.NoError(t, err)

	ctx := context.Background()

	credential := map[string]interface{}{
		"username": "test@example.com",
		"password": "secret123",
	}

	// 多次加密同样的数据，结果应该不同（因为使用随机nonce）
	encrypted1, err := vault.Encrypt(ctx, credential)
	require.NoError(t, err)

	encrypted2, err := vault.Encrypt(ctx, credential)
	require.NoError(t, err)

	// 加密结果应该不同
	assert.NotEqual(t, encrypted1["encrypted_data"], encrypted2["encrypted_data"])

	// 但解密结果应该相同
	decrypted1, err := vault.Decrypt(ctx, encrypted1)
	require.NoError(t, err)

	decrypted2, err := vault.Decrypt(ctx, encrypted2)
	require.NoError(t, err)

	assert.Equal(t, credential, decrypted1)
	assert.Equal(t, credential, decrypted2)
	assert.Equal(t, decrypted1, decrypted2)
}
