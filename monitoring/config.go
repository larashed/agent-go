package monitoring

// Config holds metric collection configuration
type Config struct {
	AppBucketLimit                 int
	AppBucketNotFillingSeconds     int
	ServerMetricCollectionInterval int
}

// NewConfig creates a new instance of `Config`
func NewConfig(
	appBucketLimit int,
	appBucketNotFillingSeconds int,
	serverMetricCollectionInterval int,
) *Config {
	return &Config{
		appBucketLimit,
		appBucketNotFillingSeconds,
		serverMetricCollectionInterval,
	}
}
