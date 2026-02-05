package health_checker

import (
	"net/http"
	"net/url"
	errors "reverse_proxy/CustomErrors"
	"reverse_proxy/core/load_balancer"
	"time"
)

type HealthChecker struct {
	client *http.Client
	healthCheckFreq *time.Duration
}

// Constructor fot HealthChecker struct
func NewHealthChecker(timeout time.Duration, healthCheckFreq *time.Duration) (*HealthChecker, error) {
	var hc = HealthChecker{}

	// Defining the client with finite timeout so that slow servers are treated as DOWN
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
	resp, err := hc.client.Head(url)
	if err != nil {
		// if the server is down the error will not be nil
		return 0, 0 * time.Second, errors.ServerDownError
	}
	// For a HEAD request, the body is empty, so no need to consume it, but still must be closed
	resp.Body.Close()
	// Calculating the response time
	duration := time.Since(start)

	return resp.StatusCode, duration, nil
}

// Method to ping a backend and update its state
func (hc *HealthChecker) PingBackendPeriodically(backend *load_balancer.Backend) {
	for {
		statusCode, responseTime, err := hc.ping_server(*backend.URL)

		backend.Alive = (statusCode != 0) && (err == nil)	// status code in convention cannot be 0
		backend.LastResponseTime = responseTime

		// waiting the time definged in the frequency
		time.Sleep(*hc.healthCheckFreq)
	}
}

// Method to ping a Load Balancer and update its state
func (hc *HealthChecker) PingLoadBalancerPeriodically(lb load_balancer.LoadBalancer) {
	var backend *load_balancer.Backend
	// Repeat forever
	for {
		// Pinging all the backends in the load balancer
		for i := range lb.GetBackendsNum() {
			backend = lb.GetBackend(i)
			statusCode, responseTime, err := hc.ping_server(*backend.URL)

			backend.UpdateStatus((statusCode != 0) && (err == nil))	// status code in convention cannot be 0
			backend.UpdateResponseTime(responseTime)
		}

		// waiting the time definged in the frequency
		time.Sleep(*hc.healthCheckFreq)
	}
}