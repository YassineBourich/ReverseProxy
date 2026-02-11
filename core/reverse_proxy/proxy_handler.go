package reverse_proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	cerrors "reverse_proxy/CustomErrors"
	"reverse_proxy/core/load_balancer"
	"reverse_proxy/core/logging"
	ratelimiter "reverse_proxy/core/rate_limiter"
	"time"
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
		return nil, fmt.Errorf("%w: %w\n", cerrors.ProxyHandlerConstErr, err)
	}

	err = json.Unmarshal(conf_file, &p.Config)
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", cerrors.ProxyHandlerConstErr, err)
	}

	// Resolving Loggin strategy
	if p.Config.LoggingEnabled {
		p.LoggingContext = &logging.FileLogger{}
	} else {
		p.LoggingContext = &logging.NoLogger{}
	}
	err = p.LoggingContext.Init()
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", cerrors.ProxyHandlerConstErr, err)
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
				http.Error(w, cerrors.HttpError(http.StatusTooManyRequests).Error(), http.StatusTooManyRequests)
				return
			}
		}

		next_handler_func(w, r)
	}
}

// Method to get the backend according to the chosen strategy
func (p *ProxyHandler) get_backend() (*load_balancer.Backend, error) {
	// resolving the load balancing strategy from the configuration file and getting the valid backend
	var backend *load_balancer.Backend
	switch p.Config.Strategy {
	case "round-robin":	
		backend = p.LoadBalancer.GetNextValidPeer()
	case "least-conn":
		backend = p.LoadBalancer.LeastConnValidPeer()
	default:
		log.Printf("Internal Configuration Error: unsupported strategy '%s'", p.Config.Strategy)
		return nil, cerrors.UnsupportedStrategyErr
	}
	// if no backend is available, handle the error
	if backend == nil {
		return nil, cerrors.BackendNotFound
	}
	return backend, nil
}

// proxyHTTP Method essential for reverse proxy
func (p *ProxyHandler) proxy_http(w http.ResponseWriter, r *http.Request) {
	var server *load_balancer.Backend

	if !p.Config.StickySessionEnabled {
		backend, err := p.get_backend()
		if err != nil {
			if errors.Is(err, cerrors.BackendNotFound) {
				// Write error status code: Service Unavailable 503
				http.Error(w, cerrors.HttpError(http.StatusServiceUnavailable).Error(), http.StatusServiceUnavailable)
			} else {
				http.Error(w, cerrors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			}
			return
		}
		server = backend
	} else {
		cookie, err := r.Cookie("reverse_proxy_backend")
		if err != nil {
			backend, err2 := p.get_backend()
			if err2 != nil {
				if errors.Is(err2, cerrors.BackendNotFound) {
					// Write error status code: Service Unavailable 503
					http.Error(w, cerrors.HttpError(http.StatusServiceUnavailable).Error(), http.StatusServiceUnavailable)
				} else {
					http.Error(w, cerrors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
				}
				return
			}
			server = backend
			// If the sticky session is enabled so the err was not nil and the sticky was not found
			// Therefore this code sets the cookie with the backend
			if p.Config.StickySessionEnabled {
				http.SetCookie(w, &http.Cookie{
					Name: "reverse_proxy_backend", 
					Value: server.URL.String(), 
					Path: "/", 
					HttpOnly: true,
					MaxAge: 3600,		// Expiration in 2 hours
				})
			}
		} else {
			// If the Sticky Session was enabled and the backend was found in the cookie, use it to find the backend in the load balancer
			server = p.LoadBalancer.FindBackendByURL(cookie.Value)
			if server == nil {
				// If the server was not found, choose another one
				backend, err2 := p.get_backend()
				if err2 != nil {
					if errors.Is(err2, cerrors.BackendNotFound) {
						// Write error status code: Service Unavailable 503
						http.Error(w, cerrors.HttpError(http.StatusServiceUnavailable).Error(), http.StatusServiceUnavailable)
					} else {
						http.Error(w, cerrors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
					}
					return
				}
				server = backend
				// If the sticky session is enabled so the server not found
				// set the cookie with the backend
				if p.Config.StickySessionEnabled {
					http.SetCookie(w, &http.Cookie{
						Name: "reverse_proxy_backend", 
						Value: server.URL.String(), 
						Path: "/", 
						HttpOnly: true,
						MaxAge: 3600,		// Expiration in 2 hours
					})
				}
			}
		}
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