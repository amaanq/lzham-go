package lzham

func clamp(value, low, high uint32) uint32 {
	if value < low {
		return low
	} else if value > high {
		return high
	}
	return value
}

func minimum(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func LZHAM_MIN(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func LZHAM_MAX(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func is_power_of_2(x uint64) bool {
	return x != 0 && (x&(x-1)) == 0
}
