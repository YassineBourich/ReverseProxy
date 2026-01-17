package load_balancer

import (
	"net/url"
)

type LoadBalancer interface {
	GetNextValidPeer() *Backend
	LeastConnValidPeer() *Backend
	AddBackend(backend *Backend)
	SetBackendStatus(uri *url.URL, alive bool)
}