package temporal

import (
	"fmt"
	"strings"
)

// formatYear is four digits for 0 to 9999 and a sign with six digits otherwise, as ISO 8601's
// extended years have it.
func formatYear(y int64) string {
	if y >= 0 && y <= 9999 {
		return fmt.Sprintf("%04d", y)
	}
	if y < 0 {
		return fmt.Sprintf("-%06d", -y)
	}
	return fmt.Sprintf("+%06d", y)
}

// formatClock is HH:MM:SS with the fraction to its last nonzero digit, omitted when zero.
func formatClock(h, m, s, nano int) string {
	out := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	if nano != 0 {
		out += "." + strings.TrimRight(fmt.Sprintf("%09d", nano), "0")
	}
	return out
}

// formatOffset is ±HH:MM, with :SS when the offset has seconds.
func formatOffset(sec int) string {
	sign := "+"
	if sec < 0 {
		sign = "-"
		sec = -sec
	}
	out := fmt.Sprintf("%s%02d:%02d", sign, sec/3600, sec/60%60)
	if sec%60 != 0 {
		out += fmt.Sprintf(":%02d", sec%60)
	}
	return out
}
