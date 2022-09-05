package main

import (
	"errors"
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
