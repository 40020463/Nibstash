package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var encryptKey []byte

// InitCrypto 初始化加密模块
func InitCrypto(key string) error {
	if len(key) != 32 {
		return errors.New("encrypt key must be 32 bytes")
	}
	encryptKey = []byte(key)
	return nil
}

// Encrypt 使用 AES-GCM 加密字符串
func Encrypt(plaintext string) (string, error) {
	if len(encryptKey) == 0 {
		return "", errors.New("encrypt key not initialized")
	}

	block, err := aes.NewCipher(encryptKey)
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

// Decrypt 使用 AES-GCM 解密字符串
func Decrypt(ciphertext string) (string, error) {
	if len(encryptKey) == 0 {
		return "", errors.New("encrypt key not initialized")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
