package reverse_proxy

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
	"os"
	"encoding/json"
	"fmt"
	"time"
)

type ProxyHandler struct {
	Config ProxyConfig
	LoadBalancer load_balancer.LoadBalancer
	ProxyCore ReverseProxyCore
}

func NewProxyHandler(LoadBalancer load_balancer.LoadBalancer, conf_file_name string) (*ProxyHandler, error) {
	var p = ProxyHandler{}
	p.ProxyCore = *NewReverseProxyCore(2 * time.Second)
	p.LoadBalancer = LoadBalancer

	conf_file, err := os.ReadFile(conf_file_name)
	if err != nil {
		return nil, fmt.Errorf("Proxy Configuration reading json flie error: %w\n", err)
	}

	err = json.Unmarshal(conf_file, &p.Config)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var server *load_balancer.Backend
	switch p.Config.Strategy {
	case "round-robin":
		server = p.LoadBalancer.GetNextValidPeer()
	case "least-conn":
		server = p.LoadBalancer.LeastConnValidPeer()
	default:
		fmt.Printf("Unsupported load balancer strategy")
		return
	}
	if server == nil {
		// Write error status code
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("503 service unavailable\n"))
		return
	}
	server.AddRequest()
	defer server.RequestDone()
	p.ProxyCore.ForwardRequest(w, r, *server.URL)
}