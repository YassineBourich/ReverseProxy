package reverse_proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

/*
Function to forward the request comming from the client
to another server and send back the response
*/
func ForwardRequest(w http.ResponseWriter, r *http.Request, server url.URL) error {
	// Resolving traget url
	var target = server
	target.Path = r.URL.Path
	target.RawQuery = r.URL.RawQuery

	fmt.Println(target)
	// Creating new request 
	req, err := http.NewRequest(r.Method, target.String(), r.Body)

	if err != nil {
		return err
	}

	// Copy headers from original request
	req.Header = r.Header.Clone()

	// Set correct Host for backend
	req.Host = server.Host

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	req.Header.Set("X-Forwarded-For", ip)

	// Performing the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	// Send back the response
	defer res.Body.Close()

	// Copy response headers
	for key, values := range res.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(res.StatusCode)

	// Copy body
	_, err = io.Copy(w, res.Body)
	return err
}