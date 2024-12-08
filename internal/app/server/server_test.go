package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	config "github.com/ry461ch/metric-collector/internal/config/server"
)

func TestBase(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	privateKeyPath := "/tmp/private.test"
	privateKeyFile, _ := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer privateKeyFile.Close()
	privateKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	privateKeyFile.Write(privateKeyBytes)

	cfg := config.New()
	cfg.CryptoKey = privateKeyPath
	server := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		server.Run(ctx)
	}()
	time.Sleep(2 * time.Second)
}
