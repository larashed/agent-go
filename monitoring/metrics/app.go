package metrics

// AppMetric application metric
type AppMetric struct {
	record string
}

// NewAppMetric creates a new `AppMetric`
func NewAppMetric(record string) *AppMetric {
	return &AppMetric{record}
}

// String returns `AppMetric` string representation
func (am *AppMetric) String() string {
	return am.record
}
