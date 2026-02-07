package main

import (
	"fmt"
	"net/http"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/reverse_proxy"
	"reverse_proxy/health_checker"
	"time"
	"reverse_proxy/admin_api"
	"path/filepath"
)


func main() {
	// Defining the load balander and proxy handler
	var LB, _ = load_balancer.NewServerPool(filepath.Join("config", "backends.json"))
	var proxy_handler, _ = reverse_proxy.NewProxyHandler(20 * time.Second, LB, filepath.Join("config", "proxy.json"))

	// Defining the proxy health checker
	hc, _ := health_checker.NewHealthChecker(time.Second, &proxy_handler.Config.HealthCheckFreq)

	// Executing the health checker in a separate goroutine (thread)
	go hc.PingLoadBalancerPeriodically(LB)
	
	// Run the proxy admin server on a separate goroutine on port 8079 with pointer to the load balancer
	go adminapi.ProxyAdmin(":8079", LB)
	
	// Running the core reverse proxy server with port provided in the configuration file
	reverse_proxy_server := &http.Server{
		Addr:         fmt.Sprintf(":%d", proxy_handler.Config.Port),
		Handler:      proxy_handler,
	}
	reverse_proxy_server.ListenAndServe()
}