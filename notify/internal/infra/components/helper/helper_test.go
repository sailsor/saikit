package helper

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	test := `{"aa":11,"bb":"22"}{"cc":345,"ff":76}`
	s1, s2 := Parse(test)
	fmt.Println(s1)
	fmt.Println(s2)
}
