package storage

import "fmt"

// Config contains the settings required to create a storage backend.
type Config struct {
	Endpoint   string
	Region     string
	AccessKey  string //nolint:gosec // G117: not a hardcoded credential
	SecretKey  string
	Bucket     string
	PublicBase string
}

// Validate returns an error if any required field is missing.
func (c Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("storage endpoint is required")
	}
	if c.Bucket == "" {
		return fmt.Errorf("storage bucket is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("storage access key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("storage secret key is required")
	}
	return nil
}
