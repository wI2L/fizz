package test_types_extra

import "github.com/wI2L/fizz/testdata/test_types"

type W struct {
	C, D int
}

type D struct {
	W  W
	Wd test_types.W
}
