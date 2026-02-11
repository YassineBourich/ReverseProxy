Author: Yassine BOURICH

# Reverse Proxy

The reverse proxy is a high-performance and a stable server standing in front of backends, providing robust request forwarding, intelligent load balancing, and administrative control.

## Configuration files

All system settings are stored as ```json``` files within the ```config/``` directory:
- **backends.json**: contains an array of objects, each represents a backend:
```json
{
    "url": "http://backend",
    "alive": true,
    "current_connections": 8
},
```
- **admin.json**: contains the username and the password to login in the administration of the reverse proxy. The password should be hashed -Bcrypt- before being stored in the file.

- **proxy.json**: The main configuration hub used to toggle features, set timeouts, and define system behavior.

## Load Balancer
A load balancer is an entity that equilibrate the traffic distribution on backends. This implementation supports two primary strategies, configurable via ```proxy.json```:
- **round-robin**: the load balancer alternate over the backends circularly.
- **least-conn**: the load balancer searches for the backend that received the least number of requests.

In ```proxy.json```:
```json
"strategy": "least-conn", // or "round-robin"
```

When picking some backend according to the chosen load balancing strategy, the load balancer returns an active (**Alive**) backend, non active backends are ignored. If no active backend is present in the server pool, the proxy returns the error with status code 503 (Service unavailable).

## Reverse Proxy Core
The Core performs the heavy lifting of request forwarding. It maps incoming traffic to backend paths seamlessly.

For instance, if your proxy is at ```localhost:8080``` and your backend provides a service at ```/api/data```, a user request to ```http://localhost:8080/api/data``` is automatically intercepted, cleaned, and forwarded to the chosen backend. 

## Health Checker
The health checker is a crucial element in the reverse proxy that allows the proxy to monitor the status of backends. The process of health checking is controlled by a frequency that is adjusted in the ```proxy.json``` configuration file:
```json
"health_check_frequency": "3s",
```

## Admin API
The administration system for this reverse proxy allows the user to monitor and control the server pool. The openapi describing the Admin API is available in ```doc/admin_api.yaml```.

The Admin API involves the following endpoints:
- ```GET /status```: allows the user to monitor the status of the server pool (total backends and number of active backends) as well as the status of each backend in the server pool.
- ```POST /backends```: allows the user to add a backend to the server pool.
- ```DELETE /backends```: allows the user to remove a backend to the server pool.
- ```POST /login```: provids a login interface for admin. Not every user can access to administration and use the Admin API.

To facilitate the interaction with this API and to ensure consistency between the login and the usage of the administration endpoints, the administration endpoint frontend can be accessed at:

- login page: ```http://localhost:8079/administration-login```.
- admin page: ```http://localhost:8079/administration```.

*Use https if ssl is enabled in ```proxy.json```.*

The default admin credentials are:
- username: admin
- password: admin

Modifying these credentials can be done by changing the values in ```admin.json```. The password must be hashed -Bcrypt- before being put there.

## Features
In addition to the essential reverse proxy functionalities, some extra features were added in order to make the reverse proxy more useful and flexible while maintaining it generic.

### Context handling
The reverse proxy handles the context in the following method: if the timeout of the request is exceeded or the connection is reset, the proxy send a signal in the request's context. To benefit from this, the backends must detect this signal, finishing the work as soon as it is triggered in order not to waste computation power in a response that will be dropped.

An example of handling the request context in the server side:
```go
func handle_request(_ http.ResponseWriter, _ *http.Request, errCh chan<- error, result chan<- []byte) {
	defer close(result)
	defer close(errCh)

	// ---- Simulating a long process ----
	time.Sleep(10 * time.Second)
	// User data
	users := []struct {
		Name     string
		Password string
	}{
		{"Yassine", "8888"},
		{"Ahmed", "1234"},
	}

	// Marshalling data in json format
	data, err := json.Marshal(users)
	if err != nil {
		// Send error through channel
		errCh <- err
		return
	}

	// Send data through channel
	result <- data
}
```
```go
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("_________________________________________________")
	ctx := r.Context()

	// Result channel
	result := make(chan []byte)
	errCh := make(chan error)

	// Start long-running process
	go handle_request(w, r, errCh, result)

	// Wait for either completion or cancellation
	select {
	case <-ctx.Done():
		fmt.Println("Request canceled")
		return

	case err := <-errCh:
		if err != nil {
			log.Println(err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

	case data := <-result:
		fmt.Println("Pass!")
		w.Write(data)
	}
}
```

### Logging
Logging is an optional feature that can be enabled in ```proxy.json```:
```json
"logging_enabled": true,
```

When enabled, it allows logging every event happens in the reverse proxy at each request, recording any potential technical or security issue in the proxy.

### Rate Limiter
The rate limiter is an integrated feature that allows limiting the rate of request per minute per client IP. It can be enabled and configured in ```proxy.json```:
```json
"rate_limiter": {"enabled": true, "max_requests_per_minute": 50},
```

### Panic Recovery
Panic recovery is a feature that allows the reverse proxy to recover from a panic if a bug occured during execution. It can be enabled in ```proxy.json```:
```json
"panic_recovery": true,
```

### Sticky Session
This a crucial element for applications where a user is communicating with a backend within a session. When a session is created between a user and a specific backend, the proxy cannot just use the load balancing strategies to forward the user's requests to another backend that does not know about that session. This may be the case for login applications or cloud based environments where the user must stick to a specific backend after the first handshake and not jumping over in the server pool.

If the backend to which a user sticked to was shutdown or removed from the server pool by an administrator, then the reverse proxy forwards the request to a new backend according the load balancing strategy and it sticks to it.

This important feature can be applied by configuring the file ```proxy.json```, exactly this part:
```json
"sticky_session_enabled": false,
```

### SSL/TLS
To run the reverse proxy over ssl using the protocol https, two files are required: the key and the ssl certificate. They can be gotten either from a trusted Certificate Authority (CA) or self-signed using openSSL by running this command:

```bash
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=localhost"
```

This feature can be configured in ```proxy.json``` by specifying the relative path to both the key file and certificate file:

```json
"ssl": {
	"enabled": true,
	"ssl-certificate": "cert.pem",
	"ssl-key": "key.pem"
}
```

It worth noting that when the connection over SSL is enabled, the requests to the reverse proxy have to be done using https: ```https://localhost:8080```

Also, the administration utility will be accessible at:
- login page: ```https://localhost:8079/administration-login```.
- admin page: ```https://localhost:8079/administration```.

## Conclusion
This Reverse Proxy is engineered to serve as a high-performance entry point for distributed architectures. By bridging the gap between raw network forwarding and intelligent traffic management, it provides a stable infrastructure layer that prioritizes resource efficiency and system resilience.

The implementation of Context-aware propagation and Non-blocking concurrency patterns ensures that backend resources are never wasted on orphaned requests, while the Multi-strategy Load Balancer allows the system to adapt to varying traffic profiles. With integrated TLS Termination, Rate Limiting, and Panic Recovery, the proxy does not merely route traffic—it protects the integrity of the entire backend ecosystem.