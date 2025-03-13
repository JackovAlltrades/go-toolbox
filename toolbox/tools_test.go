package toolbox

import "testing"

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(7)
	if len(s) != 7 {
		t.Error("the length of random string WRONG")
	}
}
