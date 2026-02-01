package load_balancer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"
	errors "reverse_proxy/CustomErrors"
)

type Backend struct {
	URL *url.URL `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns int64 `json:"current_connections"`
	LastResponseTime time.Duration `json:"last_response_time"`
	mux sync.RWMutex
}

// Auxilary backend to define custom UnmarshalJSON
type aux_backend struct {
	URL string `json:"url"`
	Alive bool `json:"alive"`
	CurrentConns int64 `json:"current_connections"`
	LastResponseTime time.Duration `json:"last_response_time"`
}

// Method to define custom Unmarshaling approche
// We have to unmarshal a url into a string then parse it
func (b *Backend) UnmarshalJSON(data []byte) error {
	aux := aux_backend{}

	// Unmarshal the data to aux and check for error
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("%w: %w\n", errors.BackendUnmarshalErr, err)
	}

	// Parsing the url
	parsed_url, err := url.Parse(aux.URL)
	if err != nil {
		return fmt.Errorf("%w: %w\n", errors.BackendUrlParsingErr, err)
	}

	// Assigning value to the original struct
	b.URL = parsed_url
	b.Alive = aux.Alive
	b.CurrentConns = aux.CurrentConns
	return nil
}

// Method to define custom Marshaling approche
// We have to Marshal a url as a string
func (b *Backend) MarshalJSON() ([]byte, error) {
	aux := aux_backend{}
	aux.Alive = b.Alive
	aux.CurrentConns = b.CurrentConns
	aux.LastResponseTime = b.LastResponseTime
	aux.URL = b.URL.String()

	// Marshaling aux into a byte array
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, fmt.Errorf("%w: %w\n", errors.BackendMarshalErr, err)
	}

	return data, nil
}

// Thread-safe counter incrementing
func (b *Backend) AddRequest() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.CurrentConns++
}

// Thread-safe counter decrementing
func (b *Backend) RequestDone() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.CurrentConns--
}

// Thread-safe status update
func (b *Backend) UpdateStatus(status bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = status
}

// Thread-safe response time update
func (b *Backend) UpdateResponseTime(responseTime time.Duration) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.LastResponseTime = responseTime
}