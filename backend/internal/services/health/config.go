package health

// Config configuration for the healthcheck system.
type Config struct {
	// Period in seconds when all health checks should run.
	Period int `yaml:"period"`
	// StartDelay is an optional delay in seconds before periodic checks start.
	StartDelay int `yaml:"startdelay"`
}
