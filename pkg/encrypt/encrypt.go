package encrypt

import (
	"crypto/hmac"
	"crypto/sha256"
)

type Encrypter struct {
	secretKey string
}

func New(secretKey string) *Encrypter {
	return &Encrypter{secretKey: secretKey}
}

func (e *Encrypter) EncryptMessage(message []byte) []byte {
	h := hmac.New(sha256.New, []byte(e.secretKey))
	h.Write(message)
	return h.Sum(nil)
}
