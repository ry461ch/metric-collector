package rsamiddleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gopkg.in/resty.v1"

	rsacomponent "github.com/ry461ch/metric-collector/pkg/rsa"
)

func mockRouter(t *testing.T, decrypter *rsacomponent.RsaDecrypter) chi.Router {
	router := chi.NewRouter()
	router.Use(DecryptRequest(decrypter))
	router.Post("/*", func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		buf.ReadFrom(req.Body)
		data := buf.Bytes()
		assert.Equal(t, []byte("Test"), data, "Invalid data")
		res.WriteHeader(http.StatusOK)
	})
	return router
}

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

	encrypter := rsacomponent.NewEncrypter(publicKeyPath)
	encrypter.Initialize(context.TODO())
	decrypter := rsacomponent.NewDecrypter(privateKeyPath)
	decrypter.Initialize(context.TODO())

	router := mockRouter(t, decrypter)
	srv := httptest.NewServer(router)
	defer srv.Close()

	reqStr := "Test"
	encryptedStr, _ := encrypter.Encrypt([]byte(reqStr))
	decryptedStr, _ := decrypter.Decrypt(encryptedStr)
	assert.Equal(t, []byte(reqStr), decryptedStr)

	client := resty.New()
	resp, _ := client.R().SetBody(encryptedStr).Post(srv.URL + "/")
	assert.Equal(t, http.StatusOK, resp.StatusCode(), "Invalid status code")

	resp, _ = client.R().SetBody(reqStr).Post(srv.URL + "/")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), "Invalid status code")
}
