package maps_test

import (
	"testing"

	"github.com/pyrrho/encoding/maps"
	"github.com/stretchr/testify/require"
)

type SimpleStruct struct {
	FieldOne   int
	FieldTwo   float64
	FieldThree string
	FieldFour  complex128
}

func TestSimpleUntaggedStruct(t *testing.T) {
	require := require.New(t)

	var (
		err              error
		actual, expected map[string]interface{}
	)

	s := SimpleStruct{
		42,
		3.14,
		"Hello World",
		complex(1, 2),
	}
	sp := &s
	var si interface{} = s
	expected = map[string]interface{}{
		"FieldOne":   42,
		"FieldTwo":   float64(3.14),
		"FieldThree": "Hello World",
		"FieldFour":  complex(1, 2),
	}

	actual, err = maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)

	actual, err = maps.Marshal(sp)
	require.NoError(err)
	require.Equal(expected, actual)

	actual, err = maps.Marshal(si)
	require.NoError(err)
	require.Equal(expected, actual)
}

func TestSimpleUntaggedStructSlice(t *testing.T) {
	require := require.New(t)

	var (
		err              error
		actual, expected []map[string]interface{}
	)

	s := []SimpleStruct{
		{
			42,
			3.14,
			"Hello World",
			complex(1, 2),
		},
		{
			2,
			6.28,
			"Goodby World",
			complex(2, 1),
		},
	}
	sp := &s
	si := []interface{}{
		s[0],
		s[1],
	}
	expected = []map[string]interface{}{
		{
			"FieldOne":   42,
			"FieldTwo":   float64(3.14),
			"FieldThree": "Hello World",
			"FieldFour":  complex(1, 2),
		},
		{
			"FieldOne":   2,
			"FieldTwo":   6.28,
			"FieldThree": "Goodby World",
			"FieldFour":  complex(2, 1),
		},
	}

	actual, err = maps.MarshalSlice(s)
	require.NoError(err)
	require.Equal(expected, actual)

	actual, err = maps.MarshalSlice(sp)
	require.NoError(err)
	require.Equal(expected, actual)

	actual, err = maps.MarshalSlice(si)
	require.NoError(err)
	require.Equal(expected, actual)
}

type SimpleStructWithInterface struct {
	FieldOne int
	FieldTwo interface{}
}

func TestSimpleStructWithInterfaceMember(t *testing.T) {
	require := require.New(t)

	var (
		err                        error
		actual, expected           map[string]interface{}
		actualSlice, expectedSlice []map[string]interface{}
	)

	s := &SimpleStructWithInterface{
		42,
		"I'm an interface{}",
	}
	expected = map[string]interface{}{
		"FieldOne": 42,
		"FieldTwo": "I'm an interface{}",
	}

	actual, err = maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)

	ss := []SimpleStructWithInterface{
		{1, "One"},
		{2, "Two"},
	}
	expectedSlice = []map[string]interface{}{
		{
			"FieldOne": 1,
			"FieldTwo": "One",
		},
		{
			"FieldOne": 2,
			"FieldTwo": "Two",
		},
	}

	actualSlice, err = maps.MarshalSlice(ss)
	require.NoError(err)
	require.Equal(expectedSlice, actualSlice)
}

type SimpleStructWithTags struct {
	FieldOne   int        ``                  // undecorated
	FieldTwo   float64    `map:"-"`           // explicitly ignored
	FieldThree string     `map:"field_three"` // explicitly named
	fieldFour  complex128 `map:"field_four"`  // unexported with name (ignored)
	fieldFive  bool       ``                  // unexported sans name
}

func TestSimpleTaggedStruct(t *testing.T) {
	require := require.New(t)

	s := &SimpleStructWithTags{
		42,
		3.14,
		"Hello World",
		complex(1, 2),
		true,
	}
	expected := map[string]interface{}{
		"FieldOne":    42,
		"field_three": "Hello World",
	}
	actual, err := maps.Marshal(s)

	require.NoError(err)
	require.Equal(expected, actual)
}

type ParentStruct struct {
	AMap    map[int]int
	AStruct NestedStruct
}

type NestedStruct struct {
	AnInt  int
	AFloat float64
}

func TestNestedStructsAndMaps(t *testing.T) {
	require := require.New(t)

	s := &ParentStruct{
		map[int]int{
			1: 2,
			3: 4,
		},
		NestedStruct{5, 6.7},
	}
	expected := map[string]interface{}{
		"AMap": map[int]int{
			1: 2,
			3: 4,
		},
		"AStruct": map[string]interface{}{
			"AnInt":  5,
			"AFloat": 6.7,
		},
	}

	actual, err := maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)
}

type TopLevelStruct struct {
	AnInt  int
	WeMust // embedded
}

type WeMust struct {
	Go // embedded
}

type Go struct {
	Deeper // embedded
}

type Deeper struct {
	unexported int
	Exported   int
}

func TestSimpleEmbeddedStructs(t *testing.T) {
	require := require.New(t)

	s := &TopLevelStruct{
		42,
		WeMust{Go{Deeper{1, 2}}},
	}
	expected := map[string]interface{}{
		"AnInt":    42,
		"Exported": 2,
	}

	actual, err := maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)
}

type LevelOne struct {
	LevelTwoLeft  // embedded, with contentious field names
	LevelTwoRight // embedded, with contentious field names
}

type LevelTwoLeft struct {
	AnInt   int
	AString string
	AFloat  float64
}

type LevelTwoRight struct {
	// Accessing `LevelOne.AnInt` would cause an ambiguous selector compiler
	// error, but `map` tag means there won't be contention when marshalling.
	AnInt int `map:"AnInt"`

	LevelThree // embedded
}

type LevelThree struct {
	// `LevelThree.AString` will be shadowed by `LevelTwoLeft.AString`.
	AString string
	// `LevelThree.AFloat` will be shadowed by `LevelTwoLeft.AFloat`, despite
	// the `map` tag.
	AFloat float64 `map:"AFloat"`
}

func TestContendingEmbeddedStructs(t *testing.T) {
	require := require.New(t)

	s := &LevelOne{
		LevelTwoLeft{
			100,
			"foo",
			3.14,
		},
		LevelTwoRight{
			200,
			LevelThree{
				"bar",
				6.28,
			},
		},
	}
	require.Equal("foo", s.AString)
	require.Equal(3.14, s.AFloat)
	// Compile-time error
	// require.Equal(0, s.AnInt)
	expected := map[string]interface{}{
		"AnInt":   200,   // From LevelTwoRight
		"AString": "foo", // From LevelTwoLeft
		"AFloat":  3.14,  // From LevelTwoLeft
	}

	actual, err := maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)
}

type MarahalerParent struct {
	AnInt            int
	AnArrayIshStruct MarshalerImplementor
}

type MarshalerImplementor struct {
	AnArray  [3]int
	Constant int
}

func (mi MarshalerImplementor) MarshalMapValue() (interface{}, error) {
	return map[string]int{
		"Arr0": mi.AnArray[0] + mi.Constant,
		"Arr1": mi.AnArray[1] + mi.Constant,
		"Arr2": mi.AnArray[2] + mi.Constant,
	}, nil
}

func TestMarshalerInterface(t *testing.T) {
	require := require.New(t)

	s := &MarahalerParent{
		42,
		MarshalerImplementor{
			[3]int{1, 2, 3},
			10,
		},
	}
	expected := map[string]interface{}{
		"AnInt": 42,
		"AnArrayIshStruct": map[string]int{
			"Arr0": 11,
			"Arr1": 12,
			"Arr2": 13,
		},
	}

	actual, err := maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)
}

type DifferentTags struct {
	FieldOne   int        `map_key:"field_one"`
	FieldTwo   float64    `map_key:"field_two"`
	FieldThree string     `map_key:"field_three"`
	FieldFour  complex128 `map_key:"field_four"`
}

func TestDifferentTags(t *testing.T) {
	require := require.New(t)

	var (
		err              error
		actual, expected map[string]interface{}
	)

	s := &DifferentTags{
		42,
		3.14,
		"Hello World",
		complex(1, 2),
	}
	expected = map[string]interface{}{
		"field_one":   42,
		"field_two":   float64(3.14),
		"field_three": "Hello World",
		"field_four":  complex(1, 2),
	}

	actual, err = maps.MarshalWithConfig(s, &maps.Config{TagName: "map_key"})
	require.NoError(err)
	require.Equal(expected, actual)
}

type PossiblyNotValues struct {
	Int1  int  `map:",omitZero"`
	Int2  int  `map:",omitZero"`
	Int3  int  `map:",OmItZeRO"`
	IntP1 *int `map:",omitNil"`
	IntP2 *int `map:",omitNil"`
	IntP3 *int `map:",OMiTnIL"`
}

func TestOmitZeroNil(t *testing.T) {
	require := require.New(t)

	var (
		err              error
		actual, expected map[string]interface{}
	)

	i := 42
	s := &PossiblyNotValues{
		Int1:  2,
		Int3:  0,
		IntP1: &i,
		IntP3: nil,
	}
	expected = map[string]interface{}{
		"Int1":  2,
		"IntP1": &i,
	}

	actual, err = maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)
}

type AsValueParent struct {
	Tagged     TaggedAsValueChild `map:",value"`
	Interfaced MarshalerAsValueChild
}

type TaggedAsValueChild struct {
	AFloat float64
	ABool  bool
}
type MarshalerAsValueChild struct {
	AnInt   int
	AString string
}

func (m *MarshalerAsValueChild) MarshalMapValue() (interface{}, error) {
	return *m, nil
}

func TestStructsAsValue(t *testing.T) {
	require := require.New(t)

	var (
		err              error
		actual, expected map[string]interface{}
	)

	s := &AsValueParent{
		TaggedAsValueChild{
			3.14,
			true,
		},
		MarshalerAsValueChild{
			42,
			"Hello World",
		},
	}
	expected = map[string]interface{}{
		"Tagged": TaggedAsValueChild{
			3.14,
			true,
		},
		"Interfaced": MarshalerAsValueChild{
			42,
			"Hello World",
		},
	}

	actual, err = maps.Marshal(s)
	require.NoError(err)
	require.Equal(expected, actual)
}
