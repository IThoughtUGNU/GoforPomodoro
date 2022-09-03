package main

import "fmt"

func NiceTimeFormatting(seconds int) string {
	if seconds > 60 {
		minutes := seconds / 60
		return fmt.Sprintf("%d minutes", minutes)
	} else {
		return fmt.Sprintf("%d seconds", seconds)
	}
}
