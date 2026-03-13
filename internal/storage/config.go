package storage

// Config contains the settings required to create a storage backend.
type Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket string
	PublicBase string
}