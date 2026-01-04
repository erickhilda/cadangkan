package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	keySize = 32 // AES-256
)

// encryptor handles password encryption and decryption.
type encryptor struct {
	key []byte
}

// newEncryptor creates a new encryptor with the given key.
func newEncryptor(key []byte) *encryptor {
	return &encryptor{key: key}
}

// EncryptPassword encrypts a password and returns a base64-encoded string.
func EncryptPassword(password string) (string, error) {
	key, err := loadOrGenerateKey()
	if err != nil {
		return "", &EncryptionError{Operation: "load key", Err: err}
	}

	enc := newEncryptor(key)
	return enc.encrypt(password)
}

// DecryptPassword decrypts a base64-encoded encrypted password.
func DecryptPassword(encrypted string) (string, error) {
	key, err := loadOrGenerateKey()
	if err != nil {
		return "", &EncryptionError{Operation: "load key", Err: err}
	}

	enc := newEncryptor(key)
	return enc.decrypt(encrypted)
}

// encrypt encrypts plaintext and returns base64-encoded ciphertext.
func (e *encryptor) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", &EncryptionError{Operation: "create cipher", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &EncryptionError{Operation: "create GCM", Err: err}
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", &EncryptionError{Operation: "generate nonce", Err: err}
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts base64-encoded ciphertext.
func (e *encryptor) decrypt(encoded string) (string, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", &EncryptionError{Operation: "decode base64", Err: err}
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", &EncryptionError{Operation: "create cipher", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &EncryptionError{Operation: "create GCM", Err: err}
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", &EncryptionError{Operation: "decrypt", Err: fmt.Errorf("ciphertext too short")}
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", &EncryptionError{Operation: "decrypt", Err: err}
	}

	return string(plaintext), nil
}

// loadOrGenerateKey loads the encryption key from disk or generates a new one.
func loadOrGenerateKey() ([]byte, error) {
	keyPath, err := getKeyPath()
	if err != nil {
		return nil, err
	}

	// Try to load existing key
	if _, err := os.Stat(keyPath); err == nil {
		return os.ReadFile(keyPath)
	}

	// Generate new key
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save key with restricted permissions
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	return key, nil
}

// getKeyPath returns the path to the encryption key file.
func getKeyPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".cadangkan", ".key"), nil
}
