package shttp

// Config configuration of the http service.
type Config struct {
	// Servicename for generated certificates.
	Servicename string `yaml:"servicename"`
	// Port of the http server.
	Port int `yaml:"port"`
	// SSLPort of the https server. If 0, only http server is started.
	SSLPort int `yaml:"sslport"`
	// ServiceURL is the externally reachable base URL of this service.
	ServiceURL string `yaml:"serviceURL"`
	// DNSNames used for generated certificates.
	DNSNames []string `yaml:"dnss"`
	// IPAddresses used for generated certificates.
	IPAddresses []string `yaml:"ips"`
	// Certificate optional path to TLS certificate file.
	Certificate string `yaml:"certificate"`
	// Key optional path to TLS private key file.
	Key string `yaml:"key"`
}
