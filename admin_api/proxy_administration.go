package adminapi

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
)

func ProxyAdmin(port string, load_balancer load_balancer.LoadBalancer) {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", HandleLogin)
	mux.HandleFunc("/status", AuthenticationMiddleware(HandleStatus(load_balancer)))
	mux.HandleFunc("/backend", AuthenticationMiddleware(HandleBackends(load_balancer)))

	admin_server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	admin_server.ListenAndServe()
}