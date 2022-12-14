// This file is part of GoforPomodoro.
//
// GoforPomodoro is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoforPomodoro is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with GoforPomodoro.  If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func NiceTimeFormatting64(seconds int64) string {
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

func NiceTimeFormatting(seconds int) string {
	return NiceTimeFormatting64(int64(seconds))
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

func TimePtr(t time.Time) *time.Time {
	return &t
}

type Pair[T, U any] struct {
	First  T
	Second U
}

type EmptyOptionalError struct{}

func (_ EmptyOptionalError) Error() string {
	return "Value is empty"
}

type Optional[T any] struct {
	value   T
	isEmpty bool
}

func OptionalOf[T any](value T) (opt Optional[T]) {
	opt.isEmpty = false
	opt.value = value

	return
}

func OptionalOfNil[T any]() (opt Optional[T]) {
	opt.isEmpty = true

	return
}

func (opt Optional[T]) GetValue() (value T, err error) {
	if opt.isEmpty {
		err = EmptyOptionalError{}
	} else {
		value = opt.value
		err = nil
	}
	return
}

func (opt Optional[T]) IsEmpty() bool {
	return opt.isEmpty
}

func YesNo(value bool) string {
	if value {
		return "Yes"
	} else {
		return "No"
	}
}

func IsCapitalizedLetter(c rune) bool {
	return 'A' <= c && c <= 'Z'
}

func IsCapitalizedLetterStr(str string) bool {
	if len(str) != 1 {
		return false
	}
	for _, c := range str {
		return IsCapitalizedLetter(c)
	}
	return false
}
