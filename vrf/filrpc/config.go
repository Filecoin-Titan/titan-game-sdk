package filrpc

import (
	"time"
)

type Config struct {
	NodeURL           string
	Timeout           time.Duration
	ContractorAddress string
	PrivateKeyStr     string
}

// Option is a single titan sdk Config.
type Option func(opts *Config)

// DefaultOption returns a default set of options.
func DefaultOption() Config {
	return Config{
		Timeout: 30 * time.Second,
	}
}

// NodeURLOption set node url
func NodeURLOption(address string) Option {
	return func(opts *Config) {
		opts.NodeURL = address
	}
}

// TimeoutOption specifies a time limit for requests made by the http Client.
func TimeoutOption(timeout time.Duration) Option {
	return func(opts *Config) {
		opts.Timeout = timeout
	}
}

// ContractorAddressOption specifies a contractor address
func ContractorAddressOption(id string) Option {
	return func(opts *Config) {
		opts.ContractorAddress = id
	}
}

// PrivateKeyStrOption specifies a private key
func PrivateKeyStrOption(key string) Option {
	return func(opts *Config) {
		opts.PrivateKeyStr = key
	}
}
