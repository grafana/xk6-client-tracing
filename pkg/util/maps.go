package util

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	var n int
	for _, m := range maps {
		n += len(m)
	}

	merged := make(map[K]V, n)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}
