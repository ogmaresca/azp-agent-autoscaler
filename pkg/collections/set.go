package collections

var void struct{}

// StringSet is a map that replicates a set for strings from languages with generics
type StringSet map[string]struct{}

// Add a value to the set
func (s StringSet) Add(key string) {
	s[key] = void
}

// Contains checks if the key belongs in the map
func (s StringSet) Contains(key string) bool {
	_, exists := s[key]
	return exists
}

// IntSet is a map that replicates a set for int32s from languages with generics
type IntSet map[int]struct{}

// Add a value to the set
func (s IntSet) Add(key int) {
	s[key] = void
}

// Contains checks if the key belongs in the map
func (s IntSet) Contains(key int) bool {
	_, exists := s[key]
	return exists
}
