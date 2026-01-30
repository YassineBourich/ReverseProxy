package reverse_proxy

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
	"os"
	"encoding/json"
	"fmt"
	"log"
	"time"
	errors "reverse_proxy/CustomErrors"
)

type ProxyHandler struct {
	Config ProxyConfig
	LoadBalancer load_balancer.LoadBalancer
	ProxyCore ReverseProxyCore
}

// Proxy handler constructor
func NewProxyHandler(timeout time.Duration, LoadBalancer load_balancer.LoadBalancer, conf_file_name string) (*ProxyHandler, error) {
	var p = ProxyHandler{}
	// Creating new reverse proxy and assigning the load balancer
	p.ProxyCore = *NewReverseProxyCore(timeout)
	p.LoadBalancer = LoadBalancer

	// Reading the configuration file and unmarshal it to Config
	conf_file, err := os.ReadFile(conf_file_name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", errors.ProxyHandlerConstErr, err)
	}

	err = json.Unmarshal(conf_file, &p.Config)
	fmt.Println(err)
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", errors.ProxyHandlerConstErr, err)
	}

	return &p, nil
}

// ServeHTTP Method essential for http handler
func (p ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var server *load_balancer.Backend
	// resolving the load balancing strategy from the configuration file and getting the valid backend
	switch p.Config.Strategy {
	case "round-robin":
		server = p.LoadBalancer.GetNextValidPeer()
	case "least-conn":
		server = p.LoadBalancer.LeastConnValidPeer()
	default:
		msg := fmt.Sprintf("Internal Configuration Error: unsupported strategy '%s'", p.Config.Strategy)
		log.Println(msg)
		http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
		return
	}
	// if no backend is available, handle the error
	if server == nil {
		// Write error status code
		http.Error(w, errors.HttpError(http.StatusServiceUnavailable).Error(), http.StatusServiceUnavailable)
		return
	}
	// Increment the backend's counter, forward request then decrement it
	server.AddRequest()
	defer server.RequestDone()
	p.ProxyCore.ForwardRequest(w, r, *server.URL)
}