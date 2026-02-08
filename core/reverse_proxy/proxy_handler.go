package reverse_proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	errors "reverse_proxy/CustomErrors"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/logging"
	ratelimiter "reverse_proxy/core/rate_limiter"
	"time"
	"net"
)

// Custom response writer to record status code for middlewares
type reverse_proxy_response_writer struct {
	http.ResponseWriter
	status_code int
	backend_url string
}
// Overriding WriteHeader method from http.ResponseWriter interface
func (sr *reverse_proxy_response_writer) WriteHeader(statusCode int) {
	if sr.status_code != 0 {
		return
	}
	sr.status_code = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

type ProxyHandler struct {
	Config *ProxyConfig
	LoadBalancer load_balancer.LoadBalancer
	ProxyCore *ReverseProxyCore
	LoggingContext logging.Logger
	RateLimiter *ratelimiter.ReverseProxyRateLimiter
	handler_func http.HandlerFunc
}

// Proxy handler constructor
func NewProxyHandler(timeout time.Duration, LoadBalancer load_balancer.LoadBalancer, conf_file_name string) (*ProxyHandler, error) {
	var p = ProxyHandler{}
	// Creating new reverse proxy and assigning the load balancer
	p.ProxyCore = NewReverseProxyCore(timeout)
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

	// Resolving Loggin strategy
	if p.Config.LoggingEnabled {
		p.LoggingContext = &logging.FileLogger{}
	} else {
		p.LoggingContext = &logging.NoLogger{}
	}
	err = p.LoggingContext.Init()
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", errors.ProxyHandlerConstErr, err)
	}
	// Creating the rate_limiter if enabled in configuration
	p.RateLimiter = ratelimiter.CreateReverseProxyRateLimiter(p.Config.RateLimiter)
	// Resolving middlewares
	p.handler_func = http.HandlerFunc(
		p.recovery_middleware(
		p.logging_middleware(
		p.ratelimiter_middleware(
		p.proxy_http,
	))))
	return &p, nil
}

// ServeHTTP Method essential for handler
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.handler_func(w, r)
}

// Middlewares definition
func (p *ProxyHandler) recovery_middleware(next_handler_func http.HandlerFunc) http.HandlerFunc {
	if p.Config.PanicRecovery {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("%s %s -> RemoteAddr: %s | Err: %s", r.Method, r.URL.Path, r.RemoteAddr, err)
					
					// Return err to the client
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error: Recovery Handled"))
				}
			}()
			
			next_handler_func(w, r)
		}
	}
	
	return next_handler_func
}

func (p *ProxyHandler) logging_middleware(next_handler_func http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// starting time counter
		start := time.Now()
		// Defining a status recorder
		response_writer := &reverse_proxy_response_writer{w, http.StatusOK, "N/A"}
		// Applying the next handler function
		next_handler_func(response_writer, r)
		// logging information
		p.LoggingContext.Log(r.Method, r.URL.Path, r.RemoteAddr, response_writer.backend_url, response_writer.status_code, time.Since(start))
	}
}

func (p *ProxyHandler) ratelimiter_middleware(next_handler_func http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p.RateLimiter != nil {
			// Resolving sender ip address
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)

			// Checking the requests rate
			if !p.RateLimiter.IsRateOK(ip) {
				http.Error(w, errors.HttpError(http.StatusTooManyRequests).Error(), http.StatusTooManyRequests)
				return
			}
		}

		next_handler_func(w, r)
	}
}

// proxyHTTP Method essential for reverse proxy
func (p *ProxyHandler) proxy_http(w http.ResponseWriter, r *http.Request) {
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
		// Write error status code: Service Unavailable 503
		http.Error(w, errors.HttpError(http.StatusServiceUnavailable).Error(), http.StatusServiceUnavailable)
		return
	}
	// Recording the response writer
	if writer, ok := w.(*reverse_proxy_response_writer); ok {
		writer.backend_url = server.URL.String()
	}
	// Increment the backend's counter, forward request then decrement it
	server.AddRequest()
	defer server.RequestDone()
	p.ProxyCore.ForwardRequest(w, r, *server.URL)
}