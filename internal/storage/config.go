package storage

// Config contains the settings required to create a storage backend.
type Config struct {
	Endpoint   string
	Region     string
	AccessKey  string //nolint:gosec // G117: not a hardcoded credential
	SecretKey  string
	Bucket     string
	PublicBase string
}
