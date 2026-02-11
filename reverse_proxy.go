package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"reverse_proxy/admin_api"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/reverse_proxy"
	"reverse_proxy/health_checker"
	"time"
)

func log_config(config *reverse_proxy.ProxyConfig) {
	log.Println("________ Reverse Proxy Configuration ________")
	log.Println("Port: ", config.Port)
	log.Println("Strategy: ", config.Strategy)
	log.Println("Health checker frequency: ", config.HealthCheckFreq)
	log.Println("Logging enabled: ", config.LoggingEnabled)
	log.Println("Rate limiter: ")
	log.Println("\tEnabled: ", config.RateLimiter.Enabled)
	log.Println("\tMaximum requests per minute: ", config.RateLimiter.MaxReqPerMin)
	log.Println("Panic recovery enabled: ", config.PanicRecovery)
	log.Println("Sticky session enabled: ", config.StickySessionEnabled)
	log.Println("SSL: ")
	log.Println("\tEnabled: ", config.SSL.Enabled)
	log.Println("\tCertificate file: ", config.SSL.SSLCert)
	log.Println("\tKey file: ", config.SSL.SSLKey)
}

func main() {
	// Defining the load balander and proxy handler
	var LB, err1 = load_balancer.NewServerPool(filepath.Join("config", "backends.json"))
	if err1 != nil {
		log.Fatal("Error in instantiating the load balancer")
	}
	var proxy_handler, err2 = reverse_proxy.NewProxyHandler(20 * time.Second, LB, filepath.Join("config", "proxy.json"))
	if err2 != nil {
		log.Fatal("Error in instantiating the reverse proxy handler")
	}

	// Logging configuration
	log_config(proxy_handler.Config)

	// Defining the proxy health checker
	hc, err := health_checker.NewHealthChecker(time.Second, &proxy_handler.Config.HealthCheckFreq)
	if err != nil {
		log.Fatal("Error in instantiating the health checker")
	}

	// Executing the health checker in a separate goroutine (thread)
	go hc.PingLoadBalancerPeriodically(LB)

	// Create a goroutine for cleaning the rate limiter if enabled
	if proxy_handler.Config.RateLimiter.Enabled {
		go proxy_handler.RateLimiter.CleanRateLimiter(10 * time.Minute, 30 * time.Minute)
	}
	
	// Run the proxy admin server on a separate goroutine on port 8079 with pointer to the load balancer, and ssl
	go adminapi.ProxyAdmin(":8079", LB, &proxy_handler.Config.SSL)
	
	// Running the core reverse proxy server with port provided in the configuration file
	reverse_proxy_server := &http.Server{
		Addr:         fmt.Sprintf(":%d", proxy_handler.Config.Port),
		Handler:      proxy_handler,
	}

	// Use SSL if enabled in the configuration
	if proxy_handler.Config.SSL.Enabled {
		reverse_proxy_server.ListenAndServeTLS(proxy_handler.Config.SSL.SSLCert, proxy_handler.Config.SSL.SSLKey)
	} else {
		reverse_proxy_server.ListenAndServe()
	}
}