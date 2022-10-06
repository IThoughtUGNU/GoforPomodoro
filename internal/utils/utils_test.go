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
	"testing"
)

func TestAfterRemoveEl(t *testing.T) {
	var st_ = [...]string{"ciao", "mondo"}
	var st []string = st_[:]

	s1, err := AfterRemoveEl(st, "ciao")
	if err != nil {
		t.Fatalf("Should have not returned error.")
	}
	if len(s1) > 1 {
		t.Fatalf("Didn't delete element!")
	}

	s2, err := AfterRemoveEl(st, "mondo")
	if err != nil {
		t.Fatalf("Should have not returned error.")
	}
	if len(s2) > 1 {
		t.Fatalf("Didn't delete element!")
	}

	s3, err := AfterRemoveEl(s1, "mondo")
	s4, err := AfterRemoveEl(s2, "ciao")

	if len(s3) > 0 || len(s4) > 0 {
		t.Fatalf("s3/s4 not equal to empty slice")
	}
}

func TestNiceTimeFormatting(t *testing.T) {
	{
		ok := NiceTimeFormatting(10) == "10 seconds"
		if !ok {
			t.Fatalf("NiceTimeFormatting(10) != \"10 seconds\" ")
		}
	}

	{
		lhs := "2 minutes"
		rhs := NiceTimeFormatting(100)
		ok := lhs == rhs
		if !ok {
			t.Fatalf("error second check. Should be %s, instead it is %s", lhs, rhs)
		}
	}

	{
		lhs := "1 hour 55 minutes"
		rhs := NiceTimeFormatting(115 * 60)
		ok := lhs == rhs
		if !ok {
			t.Fatalf("error third check. Should be %s, instead it is %s", lhs, rhs)
		}
	}
}
