package extra_types

import (
	baseTypes "github.com/wI2L/fizz/openapi/test_types/base_types"
)

type W struct {
	C, D int
}

type D struct {
	Winternal W
	Wexternal baseTypes.W
}
