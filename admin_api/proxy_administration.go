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
	mux.HandleFunc("/validate-token", AuthenticationMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
	mux.HandleFunc("/administration", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\administration.html")
	})
	mux.HandleFunc("/administration-login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\login.html")
	})
	mux.HandleFunc("/login.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\login.js")
	})
	mux.HandleFunc("/administration.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\administration.js")
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\favicon.ico")
	})
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\style.css")
	})

	admin_server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	admin_server.ListenAndServe()
}