package rsa

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	publicKey := privateKey.PublicKey

	privateKeyPath := "/tmp/private.test"
	publicKeyPath := "/tmp/public.test"

	privateKeyFile, _ := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer privateKeyFile.Close()

	privateKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	privateKeyFile.Write(privateKeyBytes)

	publicKeyFile, _ := os.OpenFile(publicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer publicKeyFile.Close()

	publicKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&publicKey),
		},
	)
	publicKeyFile.Write(publicKeyBytes)

	encrypter := NewEncrypter(publicKeyPath)
	encrypter.Initialize(context.TODO())
	decrypter := NewDecrypter(privateKeyPath)
	decrypter.Initialize(context.TODO())

	reqStr := "Test"
	encryptedStr, _ := encrypter.Encrypt([]byte(reqStr))
	decryptedStr, _ := decrypter.Decrypt(encryptedStr)
	assert.Equal(t, []byte(reqStr), decryptedStr)
}

func TestInvalid(t *testing.T) {
	invalidKeyPath := "/tmp/invalid.test"

	invalidKeyFile, _ := os.OpenFile(invalidKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer invalidKeyFile.Close()
	invalidKeyFile.Write([]byte("invalid"))

	encrypter := NewEncrypter(invalidKeyPath)
	err := encrypter.Initialize(context.TODO())
	assert.Error(t, err, "Not an error")

	decrypter := NewDecrypter(invalidKeyPath)
	err = decrypter.Initialize(context.TODO())
	assert.Error(t, err, "Not an error")
}
