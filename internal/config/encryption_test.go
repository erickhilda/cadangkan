package config

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "mysecretpassword",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!#$%^&*()",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-with-many-characters-to-test-encryption-123456789",
		},
		{
			name:     "unicode password",
			password: "密码🔐パスワード",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := EncryptPassword(tt.password)
			if err != nil {
				t.Fatalf("EncryptPassword() error = %v", err)
			}

			if encrypted == "" {
				t.Error("EncryptPassword() returned empty string")
			}

			// Encrypted should be different from original
			if encrypted == tt.password {
				t.Error("EncryptPassword() returned unencrypted password")
			}

			// Decrypt
			decrypted, err := DecryptPassword(encrypted)
			if err != nil {
				t.Fatalf("DecryptPassword() error = %v", err)
			}

			// Decrypted should match original
			if decrypted != tt.password {
				t.Errorf("DecryptPassword() = %v, want %v", decrypted, tt.password)
			}
		})
	}
}

func TestEncryptDifferentOutputs(t *testing.T) {
	password := "testpassword"

	// Encrypt the same password twice
	encrypted1, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	encrypted2, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("EncryptPassword() error = %v", err)
	}

	// The encrypted values should be different (due to random nonce)
	if encrypted1 == encrypted2 {
		t.Error("EncryptPassword() produced identical outputs for same input (nonce not random)")
	}

	// Both should decrypt to the original
	decrypted1, err := DecryptPassword(encrypted1)
	if err != nil {
		t.Fatalf("DecryptPassword() error = %v", err)
	}
	if decrypted1 != password {
		t.Errorf("DecryptPassword() = %v, want %v", decrypted1, password)
	}

	decrypted2, err := DecryptPassword(encrypted2)
	if err != nil {
		t.Fatalf("DecryptPassword() error = %v", err)
	}
	if decrypted2 != password {
		t.Errorf("DecryptPassword() = %v, want %v", decrypted2, password)
	}
}

func TestDecryptInvalidData(t *testing.T) {
	tests := []struct {
		name      string
		encrypted string
	}{
		{
			name:      "invalid base64",
			encrypted: "not-valid-base64!!!",
		},
		{
			name:      "empty string",
			encrypted: "",
		},
		{
			name:      "short ciphertext",
			encrypted: "YWJj", // "abc" in base64, too short
		},
		{
			name:      "corrupted data",
			encrypted: "AAAAAAAAAAAAAAAAAAAAAA==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptPassword(tt.encrypted)
			if err == nil {
				t.Error("DecryptPassword() expected error, got nil")
			}
		})
	}
}

func TestEncryptorWithCustomKey(t *testing.T) {
	// Test with a custom key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc := newEncryptor(key)
	password := "testpassword"

	encrypted, err := enc.encrypt(password)
	if err != nil {
		t.Fatalf("encrypt() error = %v", err)
	}

	decrypted, err := enc.decrypt(encrypted)
	if err != nil {
		t.Fatalf("decrypt() error = %v", err)
	}

	if decrypted != password {
		t.Errorf("decrypt() = %v, want %v", decrypted, password)
	}
}

func TestEncryptorWithDifferentKeys(t *testing.T) {
	// Create two encryptors with different keys
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1)
	}

	enc1 := newEncryptor(key1)
	enc2 := newEncryptor(key2)

	password := "testpassword"

	// Encrypt with first key
	encrypted, err := enc1.encrypt(password)
	if err != nil {
		t.Fatalf("encrypt() error = %v", err)
	}

	// Try to decrypt with second key (should fail)
	_, err = enc2.decrypt(encrypted)
	if err == nil {
		t.Error("decrypt() with wrong key should fail")
	}
}
