package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	Enabled bool `json:"enabled"`
	MaxReqPerMin int `json:"max_requests_per_minute"`
}

type ReverseProxyRateLimiter struct {
	sync.Mutex
	visitors map[string]time.Time
	min_time_difference time.Duration
}

// Creating the Rate limiter if the rate limiter is enabled
func CreateReverseProxyRateLimiter(rl RateLimiter) *ReverseProxyRateLimiter {
	if !rl.Enabled {
		return nil
	}
	rp_rl := ReverseProxyRateLimiter{}
	rp_rl.visitors = make(map[string]time.Time)
	rp_rl.min_time_difference = time.Minute / time.Duration(rl.MaxReqPerMin)
	return &rp_rl
}

// Function to compare the request rate of the user Thread-safely
func (rp_rl *ReverseProxyRateLimiter)IsRateOK(ip string) bool {
	rp_rl.Lock()
	defer rp_rl.Unlock()
	
	last_visit, ok := rp_rl.visitors[ip]
	visit_difference := time.Since(last_visit)

	if !ok || visit_difference >= rp_rl.min_time_difference {
		rp_rl.visitors[ip] = time.Now()
		return true
	}

	return false
}

// Function to clean the map periodically Thread-safely
func (rp_rl *ReverseProxyRateLimiter)CleanRateLimiter(freq time.Duration, time_threshold time.Duration) bool {
	for {
		time.Sleep(freq)
		rp_rl.Lock()

		for ip, last_visit := range rp_rl.visitors {
			if time.Since(last_visit) > time_threshold {
				delete(rp_rl.visitors, ip)
			}
		}
		rp_rl.Unlock()
	}
}