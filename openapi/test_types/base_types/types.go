package base_types

import "time"

type (
	W struct {
		A, B string
	}
	u struct {
		S int
	}
	q  int
	ns string
	ni int
	X  struct {
		*X // ignored, recursive embedding
		*Y
		A string `validate:"required"`
		B *int
		C bool `deprecated:"true"`
		D []*Y
		E [3]*X
		F *X
		G *Y
		H map[int]*Y // ignored, unsupported keys type
		*u
		uu *u // ignored, unexported field
		q     // ignored, embedded field of non-struct type
		*Q
		*V `json:"data"`
		NS ns
		NI *ni
	}
	Y struct {
		H float32   `validate:"required"`
		I time.Time `format:"date"`
		J *uint8    `deprecated:"oui"` // invalid value, interpreted as false
		K *Z        `validate:"required"`
		N struct {
			Na, Nb string
			Nc     time.Duration
		}
		l int // ignored
		M int `json:"-"`
	}
	Z map[string]*Y
	Q struct {
		NnNnnN string `json:"nnNnnN"`
	}
	V struct {
		L int
	}
)

func (*X) TypeName() string { return "XXX" }
func (*W) Format() string   { return "wallet" }
func (*W) Type() string     { return "string" }
func (ns) Nullable() bool   { return true }
func (ni) Nullable() bool   { return false }
