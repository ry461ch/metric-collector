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
)

type Sender struct {
	lastSendMetricTime time.Time
	cfg                *config.Config
	encrypter          *encrypt.Encrypter
}

func New(encrypter *encrypt.Encrypter, cfg *config.Config) *Sender {
	return &Sender{
		lastSendMetricTime: time.Time{},
		cfg:                cfg,
		encrypter:          encrypter,
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

				reqBodyHash := s.encrypter.EncryptMessage(reqBody)
				log.Printf("body hash: %x", reqBodyHash)

				restyRequest := client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("HashSHA256", fmt.Sprintf("%x", reqBodyHash)).
					SetBody(reqBody)
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

func (s *Sender) Run(ctx context.Context, metricChannel <-chan metrics.Metric) {
	defaultTime := time.Time{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if s.lastSendMetricTime == defaultTime ||
				time.Duration(time.Duration(s.cfg.ReportIntervalSec)*time.Second) <= time.Since(s.lastSendMetricTime) {
				sendMetricsCtx, sendMetricsCtxCancel := context.WithTimeout(ctx, 5*time.Second)
				s.sendMetrics(sendMetricsCtx, metricChannel)
				sendMetricsCtxCancel()
				s.lastSendMetricTime = time.Now()
			}
		}
		time.Sleep(time.Second)
	}
}
