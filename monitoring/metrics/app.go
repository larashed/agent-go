package metrics

type AppMetric struct {
	record string
}

func NewAppMetric(record string) *AppMetric {
	return &AppMetric{record}
}

func (am *AppMetric) String() string {
	return am.record
}

func (am *AppMetric) Value() interface{} {
	return am.record
}

