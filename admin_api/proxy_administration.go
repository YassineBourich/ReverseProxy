package adminapi

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
)

// Reverse Proxy Administration logic
func ProxyAdmin(port string, load_balancer load_balancer.LoadBalancer) {
	// Using a Serve Multiplexer to match request URL
	mux := http.NewServeMux()

	// Login Handler and Token verification handler
	mux.HandleFunc("/login", HandleLogin)
	mux.HandleFunc("/validate-token", AuthenticationMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
	// Status handler to return the status of the reverse proxy
	mux.HandleFunc("/status", AuthenticationMiddleware(HandleStatus(load_balancer)))
	// Backends handler to control the load balancer and the server pool
	mux.HandleFunc("/backends", AuthenticationMiddleware(HandleBackends(load_balancer)))
	// Frontend of the administration to facilitate request and monitoring by serving frontend files
	mux.HandleFunc("/administration", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\administration.html")
	})
	mux.HandleFunc("/administration-login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\login.html")
	})
	mux.HandleFunc("/error.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend\\error.js")
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

	// Defining and running the administration server
	admin_server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	admin_server.ListenAndServe()
}