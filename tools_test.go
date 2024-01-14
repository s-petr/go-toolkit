package toolkit

import (
	"testing"
)

var tools Tools

func Test_RandomString(t *testing.T) {
	s := tools.RandomString(11)
	if len(s) != 10 {
		t.Error("returned string has the wrong length")
	}
}
