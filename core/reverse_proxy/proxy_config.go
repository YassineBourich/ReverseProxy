package reverse_proxy

import (
	"time"
	"encoding/json"
	"fmt"
)

type ProxyConfig struct {
	Port int `json:"port"`
	Strategy string `json:"strategy"` // e.g., "round-robin" or "least-conn"
	HealthCheckFreq time.Duration `json:"health_check_frequency"`
}

func (pg *ProxyConfig) UnmarshalJSON(data []byte) error {
	aux_config := struct {
		Port int `json:"port"`
		Strategy string `json:"strategy"`
		HealthCheckFreq string `json:"health_check_frequency"`
	}{}

	if err := json.Unmarshal(data, &aux_config); err != nil {
		return fmt.Errorf("Proxy Configuration Unmarshal json error: %w\n", err)
	}

	parsed_duration, err := time.ParseDuration(aux_config.HealthCheckFreq)
	if err != nil {
		return fmt.Errorf("Proxy Configuration parsing duration error: %w\n", err)
	}

	pg.Port = aux_config.Port
	pg.Strategy = aux_config.Strategy
	pg.HealthCheckFreq = parsed_duration
	return nil
}