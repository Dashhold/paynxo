package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// EncryptPassword encrypts a plaintext password using AES-GCM with a key
// derived from the TOKEN_SECRET environment variable. This allows SuperAdmin
// to view the original password later via DecryptPassword.
func EncryptPassword(plaintext string) (string, error) {
	secret := os.Getenv("TOKEN_SECRET")
	if secret == "" {
		return "", errors.New("TOKEN_SECRET not set")
	}

	// Use first 32 bytes of secret as AES-256 key (pad if needed)
	key := []byte(secret)
	if len(key) < 32 {
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts an encrypted password string back to plaintext.
func DecryptPassword(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	secret := os.Getenv("TOKEN_SECRET")
	if secret == "" {
		return "", errors.New("TOKEN_SECRET not set")
	}

	// Use first 32 bytes of secret as AES-256 key (pad if needed)
	key := []byte(secret)
	if len(key) < 32 {
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
