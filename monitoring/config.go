package monitoring

type Config struct {
	AppBucketLimit                 int
	AppBucketNotFillingSeconds     int
	ServerMetricCollectionInterval int
}

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
