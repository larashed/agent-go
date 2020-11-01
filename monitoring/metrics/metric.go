package metrics

// Metric defines a collectable metric
type Metric interface {
	String() string
}
