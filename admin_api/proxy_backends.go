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
		// Unmarshalling the request body into aux
		aux := struct {
			Url string `json:"url"`
		}{}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&aux); err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}
		// Parsing the url and populate backend.URL
		var err error
		backend.URL, err = url.Parse(aux.Url)
		if err != nil {
			http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
			return
		}
		switch r.Method {
		// If the method is POST, add backend to the load balancer and return status code 201
		case http.MethodPost:
			err := lb.AddBackend(&backend)
			if err != nil {
				http.Error(w, errors.HttpError(http.StatusInternalServerError).Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)

		// If the method is DELETE, remove backend from the load balancer (if found) and return status code 204
		case http.MethodDelete:
			err := lb.RemoveBackend(&backend)
			if err != nil {
				http.Error(w, errors.HttpError(http.StatusNotFound).Error(), http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		default:
			// If the request's method is niether POST not DELETE return error with status code 405
			http.Error(w, errors.HttpError(http.StatusMethodNotAllowed).Error(), http.StatusMethodNotAllowed)
		}
	}
}