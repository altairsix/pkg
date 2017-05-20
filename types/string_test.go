package types_test

import (
	"sort"
	"testing"

	"github.com/altairsix/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestStringSet_Len(t *testing.T) {
	set := types.StringSet{}
	set.Add("a", "b", "c")
	assert.Equal(t, 3, set.Len())
}

func TestStringSet_IsZero(t *testing.T) {
	set := types.StringSet{}
	assert.True(t, set.IsZero())
}

func TestStringSet_Contains(t *testing.T) {
	set := types.StringSet{}
	set.Add("a")
	assert.True(t, set.Contains("a"))
}

func TestStringSet_Add(t *testing.T) {
	set := types.StringSet{}
	set.Add("a", "b")
	set.Add("c")
	assert.Equal(t, types.StringArray{"a", "b", "c"}, set.Array())
}

func TestStringSet_IsPresent(t *testing.T) {
	set := types.StringSet{}
	set.Add("a")
	assert.True(t, set.IsPresent())
}

func TestStringArray_Map(t *testing.T) {
	array := types.StringArray{"a", "b", "c"}
	array = array.Map(func(v string) string { return v + "!" })
	assert.Equal(t, types.StringArray{"a!", "b!", "c!"}, array)
}

func TestStringSet_Overlaps(t *testing.T) {
	a := types.StringSet{}
	a.Add("a", "b", "c")

	b := types.StringSet{}
	b.Add("c", "d", "e")

	assert.True(t, a.Overlaps(b))
}

func TestStringArray_Unique(t *testing.T) {
	array := types.StringArray{"a", "a", "b", "c", "c", "c"}
	array = array.Unique()
	assert.Equal(t, types.StringArray{"a", "b", "c"}, array)
}

func TestStringArray_Sort(t *testing.T) {
	array := types.StringArray{"a", "c", "b"}
	sort.Strings(array)
	assert.Equal(t, types.StringArray{"a", "b", "c"}, array)
}

func TestStringArray_Take(t *testing.T) {
	array := types.StringArray{"a", "b", "c"}
	array = array.Take(2)
	assert.Equal(t, types.StringArray{"a", "b"}, array)
}

func TestStringArray_Append(t *testing.T) {
	array := types.StringArray{"a", "b"}
	array = array.Append("c")
	assert.Equal(t, types.StringArray{"a", "b", "c"}, array)
}
