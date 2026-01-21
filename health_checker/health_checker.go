package health_checker

import (
	"net/http"
	"net/url"
	"time"
	"reverse_proxy/core/load_balancer"
	errors "reverse_proxy/CustomErrors"
)

type HealthChecker struct {
	client *http.Client
	healthCheckFreq *time.Duration
}

func NewHealthChecker(timeout time.Duration, healthCheckFreq *time.Duration) (*HealthChecker, error) {
	var hc = HealthChecker{}

	hc.client = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	hc.healthCheckFreq = healthCheckFreq

	return &hc, nil
}

func (hc *HealthChecker) ping_server(serverUrl url.URL) (int, time.Duration, error) {
	url := serverUrl.String()
	
	
	start := time.Now()
	resp, err := hc.client.Get(url)
	if err != nil {
		return 0, 0 * time.Second, errors.ServerDownError
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	return resp.StatusCode, duration, nil
}

func (hc *HealthChecker) PingServerPeriodically(backend *load_balancer.Backend) {
	for {
		statusCode, responseTime, err := hc.ping_server(*backend.URL)

		backend.Alive = (statusCode != 0) && (err == nil)
		backend.LastResponseTime = responseTime

		time.Sleep(*hc.healthCheckFreq)
	}
}