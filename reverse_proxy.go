package main

import (
	"net/http"
	"reverse_proxy/core"
)


func main() {
	//http.HandleFunc("/about", handler)
	var ph = core.ProxyHandler{}
	http.ListenAndServe(":8080", ph)
}