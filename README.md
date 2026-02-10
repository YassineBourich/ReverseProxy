Author: Yassine BOURICH

# Reverse Proxy

The reverse proxy is a server standing in front of backends, and playing the role of request forwarding along with load balancing over the backends.

## Configuration files

Configuration files are ```json``` files located in ```config/```:
- **backends.json**: contains an array of objects, each represents a backend with the following structure:
```json
{
    "url": "http://backend",
    "alive": true,
    "current_connections": 8
},
```
- **admin.json**: contains the username and the password to login in the administration of the reverse proxy. The password should be in hashed format.

- **proxy.json**: contains the configuration of the reverse proxy, and allows enabling or desabling some features.

## Load Balancer
A load balancer is an entity that equilibrate the load on backends. In this reverse proxy, two load balancing strategies are implemented:
- **round-robin**: the load balancer alternate over the backends circularly.
- **least-conn**: the load balancer searches for the backend that received the least number of requests.

Choosing which strategy to apply can be done in the ```proxy.json``` configuration file:
```json
"strategy": "least-conn", // or "round-robin"
```

When picking some backend according to the chosen load balancing strategy, the load balancer returns an active (**Alive**) backend, non active backends are ignored. If no active backend is present in the server pool, the proxy returns the error with status code 503 (Service unavailable).

## Reverse Proxy Core
The reverse proxy core is the part that performs the forwarding logic.

If backends provide a service on the path ```/x```, therefore the request from the user should be to the URL: ```http(s)://[proxy_ip]:[proxy_port]/x```. For example, for a reverse proxy running in localhost port 8080, requests should be sent to ```http://localhost:8080/x```. The request is then forwarded to some backend chosen according to the load balancing strategy. 

## Health Checker
The health checker is a crucial element in the reverse proxy that allows the proxy to monitor the status of backends. The process of health checking is controlled by a frequency that is controlled in the configuration file:
```json
"health_check_frequency": "3s",
```

## Admin API
The administration system for this reverse proxy allows the user to monitor and control the server pool

## Features


`openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=localhost"`