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

// Шифровальщик запросов
type RsaEncrypter struct {
	secretKeyFile string
	publicKey     *rsa.PublicKey
}

// Создание инстанса шифровальщика
func NewEncrypter(secretKeyFile string) *RsaEncrypter {
	return &RsaEncrypter{secretKeyFile: secretKeyFile}
}

func (re *RsaEncrypter) Initialize(ctx context.Context) error {
	file, err := os.Open(re.secretKeyFile)
	if err != nil {
		return err
	}

	var secretKey []byte
	_, err = file.Read(secretKey)
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

func (re *RsaEncrypter) Encrypt(sourceText []byte) ([]byte, error) {
	sha256_hash := sha256.New()
	var label []byte
	encryptedText, err := rsa.EncryptOAEP(sha256_hash, rand.Reader, re.publicKey, sourceText, label)
	if err != nil {
		return nil, err
	}
	return encryptedText, nil
}

type RsaDecrypter struct {
	secretKeyFile string
	privateKey    *rsa.PrivateKey
}

// Создание инстанса шифровальщика
func NewDecrypter(secretKeyFile string) *RsaDecrypter {
	return &RsaDecrypter{secretKeyFile: secretKeyFile}
}

func (rd *RsaDecrypter) Initialize(ctx context.Context) error {
	file, err := os.Open(rd.secretKeyFile)
	if err != nil {
		return err
	}

	var secretKey []byte
	_, err = file.Read(secretKey)
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

func (rd *RsaDecrypter) Decrypt(encryptedText []byte) ([]byte, error) {
	sha256_hash := sha256.New()
	var label []byte
	decryptedText, err := rsa.DecryptOAEP(sha256_hash, rand.Reader, rd.privateKey, encryptedText, label)
	if err != nil {
		return nil, err
	}
	return decryptedText, nil
}
