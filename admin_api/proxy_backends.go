package adminapi

import (
	//"encoding/json"
	"encoding/json"
	"net/http"
	"net/url"
	errors "reverse_proxy/CustomErrors"
	"reverse_proxy/core/load_balancer"
)

// Defining closure to pass the load balancer as parameter
func HandleBackends(lb load_balancer.LoadBalancer) http.HandlerFunc {
	// Returning handler function
	return func(w http.ResponseWriter, r *http.Request) {
		var backend load_balancer.Backend
		aux := struct {
			Url string `json:"url"`
		}{}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&aux); err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}
		var err error
		backend.URL, err = url.Parse(aux.Url)
		if err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodPost:
			err := lb.AddBackend(&backend)
			if err != nil {
				http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			err := lb.RemoveBackend(&backend)
			if err != nil {
				http.Error(w, errors.HttpError(http.StatusNotFound).Error(), http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, errors.HttpError(http.StatusMethodNotAllowed).Error(), http.StatusMethodNotAllowed)
		}
	}
}