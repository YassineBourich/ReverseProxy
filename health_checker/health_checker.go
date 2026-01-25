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

// Constructor fot HealthChecker struct
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

// Method to ping a backend given its url
func (hc *HealthChecker) ping_server(serverUrl url.URL) (int, time.Duration, error) {
	url := serverUrl.String()
	
	// Sending a get request while counting the time for response
	start := time.Now()
	resp, err := hc.client.Get(url)
	if err != nil {
		// if the server is down the error will not be nil
		return 0, 0 * time.Second, errors.ServerDownError
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	return resp.StatusCode, duration, nil
}

// Method to ping a backend and update its state
func (hc *HealthChecker) PingServerPeriodically(backend *load_balancer.Backend) {
	for {
		statusCode, responseTime, err := hc.ping_server(*backend.URL)

		backend.Alive = (statusCode != 0) && (err == nil)	// status code in convention cannot be 0
		backend.LastResponseTime = responseTime

		// waiting the time definged in the frequency
		time.Sleep(*hc.healthCheckFreq)
	}
}