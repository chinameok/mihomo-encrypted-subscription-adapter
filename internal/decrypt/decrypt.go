// Package decrypt implements the encrypted subscription format used by
// FlClash. It intentionally only accepts the format documented by the
// adapter; callers must verify the subscription-encryption response header.
package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec G501 -- MD5 is part of the supplier protocol.
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrEmptyPassword  = errors.New("subscription password is empty")
	ErrInvalidBase64  = errors.New("invalid encrypted subscription base64")
	ErrInvalidCipher  = errors.New("invalid encrypted subscription ciphertext")
	ErrInvalidPadding = errors.New("invalid encrypted subscription padding")
)

// Decrypt decodes a FlClash response body and returns the plaintext YAML.
// The protocol derives the AES-128 key from the 16 raw bytes of the MD5
// digest. The digest is commonly displayed as lowercase hex, but those 32
// display characters are not the AES key.
func Decrypt(body []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, ErrEmptyPassword
	}
	encoded := strings.TrimSpace(string(body))
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidBase64, err)
	}
	if len(decoded) < aes.BlockSize*2 || (len(decoded)-aes.BlockSize)%aes.BlockSize != 0 {
		return nil, ErrInvalidCipher
	}

	digest := md5.Sum([]byte(password)) // #nosec G401 -- required by protocol.
	block, err := aes.NewCipher(digest[:])
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	iv := decoded[:aes.BlockSize]
	ciphertext := decoded[aes.BlockSize:]
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)
	plaintext, err = unpadPKCS7(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(plaintext) {
		return nil, errors.New("decrypted subscription is not valid UTF-8")
	}
	return plaintext, nil
}

func unpadPKCS7(in []byte, blockSize int) ([]byte, error) {
	if len(in) == 0 || len(in)%blockSize != 0 {
		return nil, ErrInvalidPadding
	}
	n := int(in[len(in)-1])
	if n == 0 || n > blockSize || n > len(in) {
		return nil, ErrInvalidPadding
	}
	for _, b := range in[len(in)-n:] {
		if int(b) != n {
			return nil, ErrInvalidPadding
		}
	}
	return in[:len(in)-n], nil
}
