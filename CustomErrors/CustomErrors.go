package customerrors

import (
	"errors"
	"fmt"
	"net/http"
)

// Defining custom errors
var ServerDownError = errors.New("Server is down")
var BackendUnmarshalErr = errors.New("Backend Unmarshal json error")
var BackendMarshalErr = errors.New("Backend Marshal json error")
var BackendUrlParsingErr = errors.New("Backend parsing url error")
var ServerPoolUnmarshalErr = errors.New("ServerPool Unmarshal json error")
var ProxyConfUnmarshalErr = errors.New("Proxy Configuration Unmarshal json error")
var ProxyConfDurationParsingErr = errors.New("Proxy Configuration parsing duration error")
var ProxyHandlerConstErr = errors.New("Proxy handler constructor error")
var BackendNotFound = errors.New("backend not found")
var LoggerInitError = errors.New("Logger initialization error")
var UnsupportedStrategyErr = errors.New("Unsupported backends strategy")
var HttpError = func(code int) error {
	return fmt.Errorf("%d %s", code, http.StatusText(code))
}