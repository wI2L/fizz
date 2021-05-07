package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSortParams tests that the parameters of
// an operation can be sorted using multiple
// sort functions.
func TestSortParams(t *testing.T) {
	params := []*ParameterOrRef{
		{Parameter: &Parameter{Name: "Paz", In: "path"}},
		{Parameter: &Parameter{Name: "Baz", In: "header"}},
		{Parameter: &Parameter{Name: "Bar", In: "query"}},
		{Parameter: &Parameter{Name: "Zap", In: "path"}},
		{Parameter: &Parameter{Name: "Foo", In: "query"}},
	}
	// Sort by location, then by name in ascending order.
	paramsOrderedBy(byLocation, byName).Sort(params)

	assert.Len(t, params, 5)

	assert.Equal(t, params[0].Name, "Paz")
	assert.Equal(t, params[1].Name, "Zap")
	assert.Equal(t, params[2].Name, "Bar")
	assert.Equal(t, params[3].Name, "Foo")
	assert.Equal(t, params[4].Name, "Baz")
}

func byName(p1, p2 *ParameterOrRef) bool {
	return p1.Name < p2.Name
}

func byLocation(p1, p2 *ParameterOrRef) bool {
	return locationsOrder[p1.In] < locationsOrder[p2.In]
}
