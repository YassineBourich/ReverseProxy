package reverse_proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
	"errors"
	cerrors "reverse_proxy/CustomErrors"
)

type ReverseProxyCore struct {
	transport *http.Transport
	timeout time.Duration
}

// Reverse proxy core constructor
func NewReverseProxyCore(timeout time.Duration) *ReverseProxyCore {
	var proxy = ReverseProxyCore{}

	proxy.timeout = timeout
	proxy.transport = &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &proxy
}

func (proxy ReverseProxyCore) setup_request(r *http.Request) {

}

/*
Function to forward the request comming from the client
to another server and send back the response
*/
func (proxy ReverseProxyCore) ForwardRequest(w http.ResponseWriter, r *http.Request, server url.URL) error {
	// Resolving traget url
	var target = server
	target.Path = r.URL.Path
	target.RawQuery = r.URL.RawQuery

	// Defining the request's context
	r_ctx := r.Context()

	// Adding timeout to the request's context
	r_ctx, cancel := context.WithTimeout(r_ctx, proxy.timeout)
	defer cancel()

	// Creating new request 
	req, err := http.NewRequestWithContext(r_ctx, r.Method, target.String(), r.Body)

	if err != nil {
        return err
	}

	// Copy headers from original request
	req.Header = r.Header.Clone()

	// Set correct Host for backend
	req.Host = server.Host

	// Defining the X-Forwarded-For header by the ip seen by the proxy the one sent by the user
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	req.Header.Set("X-Forwarded-For", ip)

	// Performing the request
	res, err := proxy.transport.RoundTrip(req)
	if err != nil {
		fmt.Println(err)
        if errors.Is(r_ctx.Err(), context.DeadlineExceeded) {
            http.Error(w, cerrors.HttpError(http.StatusGatewayTimeout).Error(), http.StatusGatewayTimeout)
        } else {
            http.Error(w, cerrors.HttpError(http.StatusBadGateway).Error(), http.StatusBadGateway)
        }
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