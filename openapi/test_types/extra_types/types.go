package extra_types

import (
	"github.com/wI2L/fizz/openapi/test_types/base_types"
)

type W struct {
	C, D int
}

type D struct {
	Winternal W
	Wexternal base_types.W
}
