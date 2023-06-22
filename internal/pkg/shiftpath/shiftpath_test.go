package shiftpath_test

import (
	"budgeting/internal/pkg/shiftpath"
	"testing"
)

func TestShiftPath(t *testing.T) {

	head, tail := shiftpath.ShiftPath("/")

	if head != "" || tail != "/" {
		t.Fail()
	}

	head, tail = shiftpath.ShiftPath("/foo")

	if head != "foo" || tail != "/" {
		t.Fail()
	}

	head, tail = shiftpath.ShiftPath("/foo/")

	if head != "foo" || tail != "/" {
		t.Fail()
	}

	head, tail = shiftpath.ShiftPath("/foo/baz")

	if head != "foo" || tail != "/baz" {
		t.Fail()
	}

}
