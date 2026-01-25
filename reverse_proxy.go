package main

import (
	"fmt"
	"net/http"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/reverse_proxy"
	"reverse_proxy/health_checker"
	"time"
)


func main() {
	//http.HandleFunc("/about", handler)
	var LB, _ = load_balancer.NewServerPool("config\\backends.json")
	var ph, _ = reverse_proxy.NewProxyHandler(LB, "config\\proxy.json")

	hc, _ := health_checker.NewHealthChecker(time.Second, &ph.Config.HealthCheckFreq)

	for i := range LB.Backends {
		go hc.PingServerPeriodically(LB.Backends[i])

		go func(idx int) {
			for {
				fmt.Println((*LB.Backends[idx].URL).String(), ":", LB.Backends[idx].Alive, " | ", LB.Backends[idx].LastResponseTime)
				time.Sleep(3 * time.Second)
			}
		}(i)
	}

	http.ListenAndServe(fmt.Sprintf(":%d", ph.Config.Port), ph)
}