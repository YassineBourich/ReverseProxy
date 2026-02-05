package adminapi

import (
	"encoding/json"
	"net/http"
	errors "reverse_proxy/CustomErrors"
	"reverse_proxy/core/load_balancer"
)

type ProxyStatus struct {
	TotalBackends int `json:"total_backends"`
	ActiveBackends int `json:"active_backends"`
	Backends []load_balancer.Backend `json:"backends"`
}

// Defining closure to pass the load balancer as parameter
func HandleStatus(lb load_balancer.LoadBalancer) http.HandlerFunc {
	// Returning handler function
	return func(w http.ResponseWriter, r *http.Request) {
		// Imposing using method GET
		if r.Method == http.MethodGet {
			// Constructing the status
			status := ProxyStatus{}
			status.TotalBackends = lb.GetBackendsNum()
			status.ActiveBackends = lb.GetAliveBackendsNum()
			status.Backends = make([]load_balancer.Backend, status.TotalBackends)
			for i := range status.TotalBackends {
				status.Backends[i] = *lb.GetBackend(i)
			}

			// Marshaling the status
			jsonRes, err := json.Marshal(status)
			if err != nil {
				http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
				return
			}
			// Writing the response
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonRes)
		} else {
			// If the request's method is not GET return error with status code 405
			http.Error(w, errors.HttpError(http.StatusMethodNotAllowed).Error(), http.StatusMethodNotAllowed)
		}
	}
}