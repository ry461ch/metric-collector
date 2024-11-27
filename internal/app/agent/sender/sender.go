// Module for sending metrics to server
package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/resty.v1"

	config "github.com/ry461ch/metric-collector/internal/config/agent"
	"github.com/ry461ch/metric-collector/internal/models/metrics"
	"github.com/ry461ch/metric-collector/pkg/encrypt"
	"github.com/ry461ch/metric-collector/pkg/rsa"
)

// Sender для отправки метрик на сервер
type Sender struct {
	cfg          *config.Config
	encrypter    *encrypt.Encrypter
	rsaEncrypter *rsa.RsaEncrypter
	ip           string
}

// Init Metric Sender
func New(encrypter *encrypt.Encrypter, rsaEncrypter *rsa.RsaEncrypter, cfg *config.Config, ip string) *Sender {
	return &Sender{
		cfg:          cfg,
		encrypter:    encrypter,
		rsaEncrypter: rsaEncrypter,
		ip:           ip,
	}
}

func (s *Sender) sendMetricsWorker(ctx context.Context, metricChannel <-chan metrics.Metric) func() error {
	return func() error {
		serverURL := "http://" + s.cfg.Addr.Host + ":" + strconv.FormatInt(s.cfg.Addr.Port, 10)

		client := resty.New()
		for {
			select {
			case <-ctx.Done():
				return nil
			case metric := <-metricChannel:
				metricList := []metrics.Metric{metric}

				reqBody, err := json.Marshal(metricList)
				if err != nil {
					return fmt.Errorf("can't convert model Metric to json")
				}

				restyRequest := client.R().SetHeader("Content-Type", "application/json")
				if s.ip != "" {
					restyRequest.SetHeader("X-Real-IP", s.ip)
				}
				if s.encrypter != nil {
					reqBodyHash := s.encrypter.EncryptMessage(reqBody)
					restyRequest.SetHeader("HashSHA256", fmt.Sprintf("%x", reqBodyHash))
				}
				if s.rsaEncrypter != nil {
					rsaBody, encryptErr := s.rsaEncrypter.Encrypt(reqBody)
					if encryptErr != nil {
						return fmt.Errorf("can't encrypt body")
					}
					restyRequest.SetBody(rsaBody)
				} else {
					restyRequest.SetBody(reqBody)
				}

				err = resty.Backoff(func() (*resty.Response, error) {
					return restyRequest.Post(serverURL + "/updates/")
				}, resty.Retries(4), resty.WaitTime(1), resty.MaxWaitTime(5))

				if err != nil {
					log.Println("Server is not available")
					return fmt.Errorf("server is not available")
				}
			default:
				return nil
			}
		}
	}
}

func (s *Sender) sendMetrics(ctx context.Context, metricChannel <-chan metrics.Metric) {
	log.Println("Trying to send metrics")

	wg := new(errgroup.Group)

	for w := 0; w < int(s.cfg.RateLimit); w++ {
		wg.Go(s.sendMetricsWorker(ctx, metricChannel))
	}

	if err := wg.Wait(); err != nil {
		log.Printf("Error occured while sending metrics: %s", err.Error())
		return
	}

	log.Println("Successfully send all metrics")
}

// Run metric sender: get metrics from channel and make request to metric server
func (s *Sender) Run(ctx context.Context, metricChannel <-chan metrics.Metric) {
	for {
		select {
		case <-ctx.Done():
			log.Println("sender done")
			return
		default:
		}
		sendMetricsCtx, sendMetricsCtxCancel := context.WithTimeout(ctx, 5*time.Second)
		s.sendMetrics(sendMetricsCtx, metricChannel)
		sendMetricsCtxCancel()
		time.Sleep(time.Duration(s.cfg.ReportIntervalSec) * time.Second)
	}
}
