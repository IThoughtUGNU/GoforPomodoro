package main

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func NiceTimeFormatting(seconds int) string {
	if seconds > 60*60 {
		// >1 hour
		minutes := int(math.Ceil(float64(seconds) / 60.0))
		hours := int(math.Floor(float64(minutes) / 60.0))

		minutes = minutes % 60

		var hoursTxt string
		if hours == 1 {
			hoursTxt = "hour"
		} else {
			hoursTxt = "hours"
		}

		return fmt.Sprintf("%d %s %d minutes", hours, hoursTxt, minutes)
	} else if seconds > 60 {
		// >1 minute

		minutes := int(math.Ceil(float64(seconds) / 60.0))
		return fmt.Sprintf("%d minutes", minutes)
	} else {
		// <=1 minute
		return fmt.Sprintf("%d seconds", seconds)
	}
}

func In[T comparable](element T, array []T) bool {
	found := false
	for _, v := range array {
		if element == v {
			found = true
		}
	}
	return found
}

func Contains[T comparable](array []T, element T) bool {
	return In(element, array)
}

// AfterRemove
// Make a copy of a slice without an element; will not modify original slice.
func AfterRemove[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func IndexOf[T comparable](element T, data []T) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

// AfterRemoveEl
// Make a copy of a slice `s` without the element `el`.
//
// Does not modify the slice passed in input.
//
// Returns err if `el` was not in `s`.
func AfterRemoveEl[T comparable](s []T, el T) ([]T, error) {
	index := IndexOf(el, s)

	if index == -1 {
		return s, errors.New("element was not in array")
	}

	return AfterRemove(s, index), nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
