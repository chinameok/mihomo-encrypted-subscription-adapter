package decrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec G501 -- test vector for the supplier protocol.
	"encoding/base64"
	"testing"
)

func TestDecryptSyntheticVector(t *testing.T) {
	password := "fixture-password"
	plaintext := []byte("proxies:\n  - name: fixture\n    type: http\n    server: example.invalid\n    port: 443\n")
	encoded := makeFixture(t, plaintext, password)
	got, err := Decrypt(encoded, password)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("plaintext mismatch: %q", got)
	}
}

func TestDecryptRejectsWrongPasswordAndMalformedBody(t *testing.T) {
	encoded := makeFixture(t, []byte("proxies: []\n"), "right")
	if _, err := Decrypt(encoded, "wrong"); err == nil {
		t.Fatal("expected wrong password error")
	}
	if _, err := Decrypt([]byte("%%%"), "right"); err == nil {
		t.Fatal("expected base64 error")
	}
}

func FuzzDecryptNeverPanics(f *testing.F) {
	f.Add([]byte("not-base64"), "fixture-password")
	f.Add(makeFixture(f, []byte("proxies: []\n"), "fixture-password"), "fixture-password")
	f.Fuzz(func(t *testing.T, body []byte, password string) {
		_, _ = Decrypt(body, password)
	})
}

func makeFixture(t testing.TB, plaintext []byte, password string) []byte {
	t.Helper()
	digest := md5.Sum([]byte(password)) // #nosec G401 -- protocol fixture.
	block, err := aes.NewCipher(digest[:])
	if err != nil {
		t.Fatal(err)
	}
	pad := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := append(append([]byte(nil), plaintext...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	iv := bytes.Repeat([]byte{0x42}, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
	return []byte(base64.StdEncoding.EncodeToString(append(iv, ciphertext...)))
}
