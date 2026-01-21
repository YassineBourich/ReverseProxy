package reverse_proxy

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
	"os"
	"encoding/json"
	"fmt"
)

type ProxyHandler struct {
	Config ProxyConfig
	LoadBalancer load_balancer.LoadBalancer

}

func NewProxyHandler(LoadBalancer load_balancer.LoadBalancer, conf_file_name string) (*ProxyHandler, error) {
	var p = ProxyHandler{}
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
	server := (p.LoadBalancer).LeastConnValidPeer()
	server.AddRequest()
	defer server.RequestDone()
	ForwardRequest(w, r, *server.URL)
}