package main

import (
	"fmt"
	"net/http"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/reverse_proxy"
	"reverse_proxy/health_checker"
	"time"
	"reverse_proxy/admin_api"
)


func main() {
	//http.HandleFunc("/about", handler)
	var LB, _ = load_balancer.NewServerPool("config\\backends.json")
	var proxy_handler, _ = reverse_proxy.NewProxyHandler(20 * time.Second, LB, "config\\proxy.json")

	hc, _ := health_checker.NewHealthChecker(time.Second, &proxy_handler.Config.HealthCheckFreq)

	go hc.PingLoadBalancerPeriodically(LB)

	/*for i := range LB.Backends {
		go func(idx int) {
			for {
				fmt.Println((*LB.Backends[idx].URL).String(), ":", LB.Backends[idx].Alive, " | ", LB.Backends[idx].LastResponseTime)
				time.Sleep(time.Second)
			}
		}(i)
	}*/
	
	go adminapi.ProxyAdmin(":8079", LB)
	
	reverse_proxy_server := &http.Server{
		Addr:         fmt.Sprintf(":%d", proxy_handler.Config.Port),
		Handler:      proxy_handler,
	}

	reverse_proxy_server.ListenAndServe()
}