package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	publicKey := privateKey.PublicKey

	encrypter := RsaEncrypter{
		publicKey: &publicKey,
	}
	decrypter := RsaDecrypter{
		privateKey: privateKey,
	}

	reqStr := "Test"
	encryptedStr, _ := encrypter.Encrypt([]byte(reqStr))
	decryptedStr, _ := decrypter.Decrypt(encryptedStr)
	assert.Equal(t, []byte(reqStr), decryptedStr)
}
