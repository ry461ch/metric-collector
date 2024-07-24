package encryptmiddleware

import (
	"net/http"
	"bytes"

	"github.com/ry461ch/metric-collector/pkg/encrypt"
)

type AccessChecker struct {
	secretKey		string
}

func New(secretKey string) *AccessChecker {
	return &AccessChecker{secretKey: secretKey}
}

type ResponseEncrypter struct {
	http.ResponseWriter
	encrypter *encrypt.Encrypter
}

func (re *ResponseEncrypter) Write(b []byte) (int, error) {
	bodyHash := re.encrypter.EncryptMessage(b)
	re.Header().Set("HashSHA256", string(bodyHash))
	return re.Write(b)
}

func CheckRequestAndEncryptResponse(encrypter *encrypt.Encrypter) func(http.Handler) http.Handler {
	return func (next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqHeaderHash256 := r.Header.Get("HashSHA256")
			if reqHeaderHash256 == "" {
				next.ServeHTTP(w, r)
				return
			}

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			reqBody := buf.Bytes()
			reqHash := encrypter.EncryptMessage(reqBody)

			if string(reqHash) != reqHeaderHash256 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			next.ServeHTTP(&ResponseEncrypter{encrypter: encrypter}, r)
		})
	}
}