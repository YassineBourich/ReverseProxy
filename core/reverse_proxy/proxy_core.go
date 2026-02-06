package reverse_proxy

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
	"errors"
	cerrors "reverse_proxy/CustomErrors"
	"strings"
)

type ReverseProxyCore struct {
	transport *http.Transport
	timeout time.Duration
	hopByHopHeaders []string
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

	// Defining hop-by-hop headers
	proxy.hopByHopHeaders = []string{
        "Connection",
        "Keep-Alive",
        "Proxy-Authenticate",
        "Proxy-Authorization",
        "Te",
        "Trailers",
        "Transfer-Encoding",
        "Upgrade",
    }

	return &proxy
}

func (proxy ReverseProxyCore) clean_request_headers(r *http.Request, client_ip string) {
	// deleting headers from the Connection header
	if c := r.Header.Get("Connection"); c != "" {
		// Reolving connection headers
		for _, extra := range strings.Split(c, ",") {
			// Removing them
			name := strings.TrimSpace(extra)
			if name != "" {
				r.Header.Del(name)
			}
		}
	}

	// deleting hop-by-hop headers
	for _, h := range proxy.hopByHopHeaders {
		r.Header.Del(h)
	}

	// Appending the X-Forwarded-For header by the ip seen by the proxy not the one sent by the user
	if client_ip == "" {
		return
	}

	prior := r.Header.Get("X-Forwarded-For")
	if prior != "" {
		client_ip = prior + ", " + client_ip
	}
	r.Header.Set("X-Forwarded-For", client_ip)
}

func (proxy ReverseProxyCore) returning_response(w http.ResponseWriter, res *http.Response) error {
	// Send back the response
	defer res.Body.Close()

	// deleting hop-by-hop headers
	for _, h := range proxy.hopByHopHeaders {
		res.Header.Del(h)
	}

	// Copy response headers
	for key, values := range res.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(res.StatusCode)

	// Copy body
	_, err := io.Copy(w, res.Body)
	return err
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

	// Cleaning headers of the request
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = ""
	}
	proxy.clean_request_headers(req, ip)

	// Performing the request
	res, err := proxy.transport.RoundTrip(req)
	if err != nil {
        if errors.Is(r_ctx.Err(), context.DeadlineExceeded) {
            http.Error(w, cerrors.HttpError(http.StatusGatewayTimeout).Error(), http.StatusGatewayTimeout)
        } else {
            http.Error(w, cerrors.HttpError(http.StatusBadGateway).Error(), http.StatusBadGateway)
        }
		return err
	}

	// Returning the returning the response to the client
	return proxy.returning_response(w, res)
}