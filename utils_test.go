package main

import (
	"GoforPomodoro/internal/utils"
	"testing"
)

func TestAfterRemoveEl(t *testing.T) {
	var st_ = [...]string{"ciao", "mondo"}
	var st []string = st_[:]

	s1, err := utils.AfterRemoveEl(st, "ciao")
	if err != nil {
		t.Fatalf("Should have not returned error.")
	}
	if len(s1) > 1 {
		t.Fatalf("Didn't delete element!")
	}

	s2, err := utils.AfterRemoveEl(st, "mondo")
	if err != nil {
		t.Fatalf("Should have not returned error.")
	}
	if len(s2) > 1 {
		t.Fatalf("Didn't delete element!")
	}

	s3, err := utils.AfterRemoveEl(s1, "mondo")
	s4, err := utils.AfterRemoveEl(s2, "ciao")

	if len(s3) > 0 || len(s4) > 0 {
		t.Fatalf("s3/s4 not equal to empty slice")
	}
}

func TestNiceTimeFormatting(t *testing.T) {
	{
		ok := utils.NiceTimeFormatting(10) == "10 seconds"
		if !ok {
			t.Fatalf("NiceTimeFormatting(10) != \"10 seconds\" ")
		}
	}

	{
		lhs := "2 minutes"
		rhs := utils.NiceTimeFormatting(100)
		ok := lhs == rhs
		if !ok {
			t.Fatalf("error second check. Should be %s, instead it is %s", lhs, rhs)
		}
	}

	{
		lhs := "1 hour 55 minutes"
		rhs := utils.NiceTimeFormatting(115 * 60)
		ok := lhs == rhs
		if !ok {
			t.Fatalf("error third check. Should be %s, instead it is %s", lhs, rhs)
		}
	}
}
