package util

import "fmt"

// Bytes turns a raw byte count into a short human-readable string.
// Used for process memory in the tree and details panes.
func Bytes(n uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.0f KB", float64(n)/float64(kb))
	}
	return fmt.Sprintf("%d B", n)
}

// Percent formats a 0-100 float as one decimal place.
func Percent(p float64) string {
	return fmt.Sprintf("%.1f%%", p)
}