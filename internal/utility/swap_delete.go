package utility

// SwapDelete removes element at index i from slice s by swapping it with the last element and then reslicing it
// to a smaller length. This changes the order of elements in the slice but is faster than moving all elements after
// the index one element to the front.
func SwapDelete[S ~[]E, E any](s S, i int) S {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
