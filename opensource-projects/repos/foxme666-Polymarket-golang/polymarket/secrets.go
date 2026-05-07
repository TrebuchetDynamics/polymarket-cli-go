package polymarket

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	DefaultEncKeyEnv  = "POLY_ENC_KEY"
	DefaultEncSaltEnv = "POLY_ENC_SALT"
)

// LoadEncryptedSecret decrypts a Fernet token from disk using env-provided key material.
func LoadEncryptedSecret(path string) (string, error) {
	return LoadEncryptedSecretWithEnv(path, DefaultEncKeyEnv, DefaultEncSaltEnv)
}

// LoadEncryptedSecretWithEnv decrypts a Fernet token from disk using the provided env vars.
func LoadEncryptedSecretWithEnv(path, keyEnv, saltEnv string) (string, error) {
	keyMaterial := strings.TrimSpace(os.Getenv(keyEnv))
	if keyMaterial == "" {
		return "", fmt.Errorf("missing environment variable: %s", keyEnv)
	}
	salt := os.Getenv(saltEnv)

	key, err := deriveFernetKey(keyMaterial, salt)
	if err != nil {
		return "", err
	}

	tokenBytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to read encrypted secret: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", errors.New("encrypted secret is empty")
	}

	plaintext, err := decryptFernetToken(token, key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(plaintext)), nil
}

// ResolveSecret returns the plaintext secret from env or decrypts it from an encrypted file.
// If both are unset, it returns an empty string and a nil error.
func ResolveSecret(envVar, encFileEnv string) (string, error) {
	if value := strings.TrimSpace(os.Getenv(envVar)); value != "" {
		return value, nil
	}
	if encPath := strings.TrimSpace(os.Getenv(encFileEnv)); encPath != "" {
		return LoadEncryptedSecret(encPath)
	}
	return "", nil
}

func deriveFernetKey(material, salt string) ([]byte, error) {
	normalized := strings.TrimSpace(material)
	if normalized == "" {
		return nil, errors.New("encryption material is empty")
	}

	if decoded, err := decodeURLBase64(normalized); err == nil && len(decoded) == 32 {
		return decoded, nil
	}

	if salt == "" {
		return nil, errors.New("provided encryption material is not a Fernet key and no salt was supplied")
	}

	derived := pbkdf2.Key([]byte(normalized), []byte(salt), 390000, 32, sha256.New)
	return derived, nil
}

func decryptFernetToken(token string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid Fernet key length: %d", len(key))
	}

	payload, err := decodeURLBase64(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Fernet token: %w", err)
	}
	if len(payload) < 1+8+16+32 {
		return nil, errors.New("Fernet token is too short")
	}
	if payload[0] != 0x80 {
		return nil, fmt.Errorf("unsupported Fernet token version: 0x%02x", payload[0])
	}

	signed := payload[:len(payload)-32]
	expectedMAC := payload[len(payload)-32:]

	mac := hmac.New(sha256.New, key[:16])
	mac.Write(signed)
	if !hmac.Equal(expectedMAC, mac.Sum(nil)) {
		return nil, errors.New("Fernet token signature mismatch")
	}

	iv := signed[9:25]
	cipherText := signed[25:]
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, errors.New("invalid Fernet ciphertext length")
	}

	block, err := aes.NewCipher(key[16:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	plain := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, cipherText)
	return pkcs7Unpad(plain, aes.BlockSize)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid PKCS7 padding size")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize {
		return nil, errors.New("invalid PKCS7 padding")
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return nil, errors.New("invalid PKCS7 padding")
		}
	}
	return data[:len(data)-pad], nil
}

func decodeURLBase64(input string) ([]byte, error) {
	if decoded, err := base64.URLEncoding.DecodeString(input); err == nil {
		return decoded, nil
	}
	return base64.RawURLEncoding.DecodeString(input)
}
