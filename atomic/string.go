package atomic

// String is an atomic type-safe wrapper around Value for strings.
type String struct{ v Value }

// NewString creates a String.
func NewString(str string) *String {
	s := &String{}
	if str != "" {
		s.Store(str)
	}
	return s
}

// Load atomically loads the wrapped string.
func (s *String) Load() string {
	v := s.v.Load()
	if v == nil {
		return ""
	}
	return v.(string)
}

// Store atomically stores the passed string.
// Note: Converting the string to an interface{} to store in the Value
// requires an allocation.
func (s *String) Store(str string) {
	s.v.Store(str)
}
