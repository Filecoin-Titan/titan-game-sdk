package client

const (
	EndpointMainnet  = "https://api.node.glif.io/"
	EndpointCalibnet = "https://api.calibration.node.glif.io/"
)

type Config struct {
	endpoint   string
	privateKey string
}

type Option func(opts *Config)

func EndpointOption(endpoint string) Option {
	return func(opts *Config) {
		opts.endpoint = endpoint
	}
}

func PrivateKeyOption(privateKey string) Option {
	return func(opts *Config) {
		opts.privateKey = privateKey
	}
}

func defaultConfig() Config {
	return Config{
		endpoint: EndpointCalibnet,
	}
}
