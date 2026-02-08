package ssl_tls

type SSL_TLS struct {
	Enabled bool `json:"enabled"`
	SSLCert string `json:"ssl-certificate"`
	SSLKey string `json:"ssl-key"`
}