package rsaencrypt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/ry461ch/metric-collector/pkg/logging"
)

// Шифровальщик запросов
type RsaEncrypter struct {
	secretKeyFile string
	secretKey     []byte
}

// Создание инстанса шифровальщика
func New(secretKeyFile string) *RsaEncrypter {
	return &RsaEncrypter{secretKeyFile: secretKeyFile}
}

func (re *RsaEncrypter) Initialize(ctx context.Context) {
	file, err := os.Open("file.txt")
	if err != nil {
		logging.Logger.Fatal(err)
	}

	_, err = file.Read(re.secretKey)
	if err != nil {
		logging.Logger.Fatal(err)
	}
}

func (re *RsaEncrypter) Encrypt(sourceText []byte) ([]byte, error) {
	block, _ := pem.Decode(re.secretKey)
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	sha256_hash := sha256.New()
	var label []byte
	encryptedText, err := rsa.EncryptOAEP(sha256_hash, rand.Reader, key.(*rsa.PublicKey), sourceText, label)
	if err != nil {
		return nil, err
	}
	return encryptedText, nil
}

func (re *RsaEncrypter) Decrypt(encryptedText []byte) ([]byte, error) {
	block, _ := pem.Decode(re.secretKey)
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logging.Logger.Fatal(err)
	}

	sha256_hash := sha256.New()
	var label []byte
	decryptedText, err := rsa.DecryptOAEP(sha256_hash, rand.Reader, key, encryptedText, label)
	if err != nil {
		return nil, err
	}
	return decryptedText, nil
}
