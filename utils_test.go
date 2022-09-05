package main

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
