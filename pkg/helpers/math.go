package helpers

// MaxInt returns the biggest int in the arguments
// Necessary because Go doesn't support ints for its built-in max() function
func MaxInt(arg1 int, args ...int) int {
	max := arg1
	for _, arg := range args {
		if max < arg {
			max = arg
		}
	}

	return max
}

// MaxInt32 returns the biggest int32 in the arguments
// Necessary because Go doesn't support int32s for its built-in max() function
func MaxInt32(arg1 int32, args ...int32) int32 {
	max := arg1
	for _, arg := range args {
		if max < arg {
			max = arg
		}
	}

	return max
}

// MinInt returns the smallest int in the arguments
// Necessary because Go doesn't support ints for its built-in min() function
func MinInt(arg1 int, args ...int) int {
	min := arg1
	for _, arg := range args {
		if min > arg {
			min = arg
		}
	}

	return min
}

// MaxInt32 returns the smallest int32 in the arguments
// Necessary because Go doesn't support int32s for its built-in min() function
func MinInt32(arg1 int32, args ...int32) int32 {
	min := arg1
	for _, arg := range args {
		if min > arg {
			min = arg
		}
	}

	return min
}
