package pb

const (
	// AverageMetricAge average over a one-minute period, which means the average
	// age of the metrics is in the period of 30 seconds
	AverageMetricAge float64 = 30.0

	// Decay formula for computing the decay factor for average metric age
	Decay float64 = 2 / (float64(AverageMetricAge) + 1)
)

// EWMA represents the exponentially weighted moving average of a series of numbers.
type EWMA struct {
	value float64 // The current value of the average.
}

// Add a value to the series and update the moving average.
func (e *EWMA) Add(value float64) {
	if e.value == 0 { // perhaps first input, no decay factor needed
		e.value = value
		return
	}
	e.value = (value * Decay) + (e.value * (1 - Decay))
}

// Value returns the current value of the moving average.
func (e *EWMA) Value() float64 {
	return e.value
}
