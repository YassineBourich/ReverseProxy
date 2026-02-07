package reverse_proxy

import (
	"time"
	"encoding/json"
	"fmt"
	errors "reverse_proxy/CustomErrors"
)

type ProxyConfig struct {
	Port int `json:"port"`
	Strategy string `json:"strategy"` // e.g., "round-robin" or "least-conn"
	HealthCheckFreq time.Duration `json:"health_check_frequency"`
	LoggingEnabled bool `json:"logging_enabled"`
	PanicRecovery bool `json:"panic_recovery"`
}

type aux_config struct {
	Port int `json:"port"`
	Strategy string `json:"strategy"`
	HealthCheckFreq string `json:"health_check_frequency"`
	LoggingEnabled bool `json:"logging_enabled"`
	PanicRecovery bool `json:"panic_recovery"`
}

// Defining a custom method to unmarshal data as we need to parse time duration to seconds
func (pg *ProxyConfig) UnmarshalJSON(data []byte) error {
	aux := aux_config{}

	// Unmarshaling data into aux and handling error
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("%w: %w\n", errors.ProxyConfUnmarshalErr, err)
	}

	// Parsing the time duration string and handling error
	parsed_duration, err := time.ParseDuration(aux.HealthCheckFreq)
	if err != nil {
		return fmt.Errorf("%w: %w\n", errors.ProxyConfDurationParsingErr, err)
	}

	// Assigning values to the original struct
	pg.Port = aux.Port
	pg.Strategy = aux.Strategy
	pg.HealthCheckFreq = parsed_duration
	pg.LoggingEnabled = aux.LoggingEnabled
	pg.PanicRecovery = aux.PanicRecovery
	return nil
}