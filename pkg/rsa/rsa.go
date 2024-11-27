package rsa

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
)

// Шифровальщик запросов RSA
type RsaEncrypter struct {
	secretKeyFile string
	publicKey     *rsa.PublicKey
}

// Создание инстанса шифровальщика RSA
func NewEncrypter(secretKeyFile string) *RsaEncrypter {
	return &RsaEncrypter{secretKeyFile: secretKeyFile}
}

// Инициализация шифровальщика
func (re *RsaEncrypter) Initialize(ctx context.Context) error {
	var secretKey []byte
	secretKey, err := os.ReadFile(re.secretKeyFile)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(secretKey)
	re.publicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return err
	}
	return nil
}

// Шифрование сообщения публичным ключом
func (re *RsaEncrypter) Encrypt(sourceText []byte) ([]byte, error) {
	sha256Hash := sha256.New()
	var label []byte
	encryptedText, err := rsa.EncryptOAEP(sha256Hash, rand.Reader, re.publicKey, sourceText, label)
	if err != nil {
		return nil, err
	}
	return encryptedText, nil
}

// Расшифровщик запросов RSA
type RsaDecrypter struct {
	secretKeyFile string
	privateKey    *rsa.PrivateKey
}

// Создание инстанса расшифровщика
func NewDecrypter(secretKeyFile string) *RsaDecrypter {
	return &RsaDecrypter{secretKeyFile: secretKeyFile}
}

// Инициализация расшифровщика
func (rd *RsaDecrypter) Initialize(ctx context.Context) error {
	var secretKey []byte
	secretKey, err := os.ReadFile(rd.secretKeyFile)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(secretKey)
	rd.privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	return nil
}

// Расшифровка сообщения приватным ключом
func (rd *RsaDecrypter) Decrypt(encryptedText []byte) ([]byte, error) {
	sha256Hash := sha256.New()
	var label []byte
	decryptedText, err := rsa.DecryptOAEP(sha256Hash, rand.Reader, rd.privateKey, encryptedText, label)
	if err != nil {
		return nil, err
	}
	return decryptedText, nil
}
