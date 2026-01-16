package core

import (
	"net/http"
	"net/url"
)

type ProxyHandler struct {
	config ProxyConfig
}


func (p ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var server_url = url.URL{Host: "localhost:8081"}
	ForwardRequest(w, r, server_url)
}