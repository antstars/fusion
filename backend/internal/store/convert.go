package store

// boolToInt converts Go bool to database INTEGER (0/1).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool converts database INTEGER (0/1) to Go bool.
func intToBool(i int) bool {
	return i != 0
}
