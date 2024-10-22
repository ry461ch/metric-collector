package encrypt

import (
	"crypto/hmac"
	"crypto/sha256"
)

// Шифровальщик запросов
type Encrypter struct {
	secretKey string
}

// Создание инстанса шифровальщика
func New(secretKey string) *Encrypter {
	return &Encrypter{secretKey: secretKey}
}

// Шифрование сообщения
func (e *Encrypter) EncryptMessage(message []byte) []byte {
	h := hmac.New(sha256.New, []byte(e.secretKey))
	h.Write(message)
	return h.Sum(nil)
}
