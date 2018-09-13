package null

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

// NullInt64 is a wrapper around the database/sql NullInt64 type that implements
// all of the encoding/type interfaces that sql.NullInt64 doesn't implement out
// of the box.
//
// If the NullInt64 is valid and contains 0, it will be considered non-nil, and
// zero.
type NullInt64 struct {
	sql.NullInt64
}

// Constructors

// Int64 creates a new NullInt64 based on the type and value of the given
// interface. This function intentionally sacrafices compile-time safety for
// developer convenience.
//
// If the interface is nil or a nil *Int64, the new NullInt64 will be null.
//
// If the interface is a int, an int64, or a non-nil *Int64, the new NullInt64
// will be valid, and will be initialized with the (possibly dereferenced) value
// of the interface.
//
// If the interface is any other type this function will panic.
func Int64(i interface{}) NullInt64 {
	switch v := i.(type) {
	case int64:
		return Int64From(v)
	case *int64:
		return Int64FromPtr(v)
	case int:
		return Int64From(int64(v))
	case nil:
		return NullInt64{}
	}
	panic(fmt.Errorf(
		"null.Int64: the given argument (%#v of type %T) was not of type "+
			"int, int64, *int64, or nil", i, i))
}

// Int64From creates a valid NullInt64 from i.
func Int64From(i int64) NullInt64 {
	return NullInt64{sql.NullInt64{
		Int64: i,
		Valid: true,
	}}
}

// Int64FromPtr creates a valid NullInt64 from *i.
func Int64FromPtr(i *int64) NullInt64 {
	if i == nil {
		return NullInt64{}
	}
	return Int64From(*i)
}

// Getters and Setters

// ValueOrZero returns the value of this NullInt64 if it is valid; otherwise it
// returns the zero value for a int64.
func (i NullInt64) ValueOrZero() int64 {
	if !i.Valid {
		return 0
	}
	return i.Int64
}

// Ptr returns a pointer to this NullInt64's value if it is valid; otherwise
// returns a nil pointer. The captured pointer will be able to modify the value
// of this NullInt64.
func (i *NullInt64) Ptr() *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// Set modifies the value stored in this NullInt64, and guarantees it is valid.
func (i *NullInt64) Set(v int64) {
	i.Int64 = v
	i.Valid = true
}

// Null marks this NullInt64 as null with no meaningful value.
func (i *NullInt64) Null() {
	i.Int64 = 0
	i.Valid = false
}

// Interfaces

// IsNil implements the pyrrho/encoding IsNiler interface. It will return true
// if this NullInt64 is null.
func (i NullInt64) IsNil() bool {
	return !i.Valid
}

// IsZero implements the pyrrho/encoding IsZeroer interface. It will return true
// if this NullInt64 is null or if its value is 0.
func (i NullInt64) IsZero() bool {
	return !i.Valid || i.Int64 == 0
}

// MarshalText implements the encoding TextMarshaler interface. It will encode
// this NullInt64 into its textual representation if valid, or an empty string
// otherwise.
func (i NullInt64) MarshalText() ([]byte, error) {
	if !i.Valid {
		return []byte{}, nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

// UnmarshalText implements the encoding TextUnmarshaler interface. It will
// decode a given []byte into this NullInt64, so long as the provided string
// is a valid textual representation of an int or a null. Empty strings and
// "null" will decode into a null NullInt64.
//
// If the decode fails, the value of this N will be unchanged.
func (i *NullInt64) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" || str == "null" {
		i.Valid = false
		return nil
	}
	tmp, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	i.Int64 = tmp
	i.Valid = true
	return nil
}

// MarshalJSON implements the encoding/json Marshaler interface. It will encode
// this NullInt64 into its JSON representation if valid, or 'null' otherwise.
func (i NullInt64) MarshalJSON() ([]byte, error) {
	if !i.Valid {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(i.Int64, 10)), nil
}

// UnmarshalJSON implements the encoding/json Unmarshaler interface. It will
// decode a given []byte into this NullInt64, so long as the provided []byte is
// a valid JSON representation of an int or a null.
//
// Empty strings and 'null' will both decode into a null NullInt64. JSON objects
// in the form of '{"Int64":<int>,"Valid":<bool>}' will decode directly into
// this NullInt64.
//
// If the decode fails, the value of this NullInt64 will be unchanged.
func (i *NullInt64) UnmarshalJSON(data []byte) error {
	var j interface{}
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	switch val := j.(type) {
	case float64:
		// Perform a second unmarshal, this time into an int64. This give the
		// JSON parse a change to meaningfully fail (eg. if val is a float).
		var tmp int64
		err := json.Unmarshal(data, &tmp)
		if err != nil {
			return err
		}
		i.Int64 = tmp
		i.Valid = true
		return nil
	case string:
		tmp, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		i.Int64 = tmp
		i.Valid = true
		return nil
	case map[string]interface{}:
		// If we've received a JSON object, try to decode it directly into our
		// sql.NullInt64. Return any errors that occur.
		// TODO: Make sure this, if `data` is malformed, can't affect the value
		//       of this NullInt64.
		return json.Unmarshal(data, &i.NullInt64)
	case nil:
		i.Int64 = 0
		i.Valid = false
		return nil
	default:
		return fmt.Errorf(
			"null: cannot unmarshal %T (%#v) into Go value of type "+
				"null.NullInt64",
			j, j,
		)
	}
}

// MarshalMapValue implements the pyrrho/encoding/maps Marshaler interface. It
// will encode this NullInt64 into its interface{} representation for use in a
// map[string]interface{} if valid, or return nil otherwise.
func (i NullInt64) MarshalMapValue() (interface{}, error) {
	if i.Valid {
		return i.Int64, nil
	}
	return nil, nil
}