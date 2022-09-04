package main

import (
	"fmt"
	"math"
)

func NiceTimeFormatting(seconds int) string {
	if seconds > 60 {
		minutes := int(math.Ceil(float64(seconds) / 60.0))
		return fmt.Sprintf("%d minutes", minutes)
	} else {
		return fmt.Sprintf("%d seconds", seconds)
	}
}
