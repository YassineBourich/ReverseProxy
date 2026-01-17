package reverse_proxy

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
)

type ProxyHandler struct {
	Config ProxyConfig
	LoadBalancer load_balancer.LoadBalancer

}

func (p *ProxyHandler) InitializeHandler(LoadBalancer load_balancer.LoadBalancer) {
	p.LoadBalancer = LoadBalancer
}

func (p ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := (p.LoadBalancer).LeastConnValidPeer()
	server.AddRequest()
	//defer server.RequestDone()
	ForwardRequest(w, r, *server.URL)
}