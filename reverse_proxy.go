package main

import (
	"net/http"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/reverse_proxy"
)


func main() {
	//http.HandleFunc("/about", handler)
	var LB = &load_balancer.ServerPool{}
	LB.LoadConfiguration()
	var ph = reverse_proxy.ProxyHandler{}
	ph.InitializeHandler(LB)
	http.ListenAndServe(":8080", ph)
}