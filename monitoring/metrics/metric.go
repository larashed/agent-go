package metrics

type Metric interface {
	String() string
	Value() interface{}
}
