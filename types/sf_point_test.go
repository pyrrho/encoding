package types_test

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-geom"

	"github.com/pyrrho/encoding/maps"
	"github.com/pyrrho/encoding/types"
)

var (
	// These are all OpenGIS Simple Feature representations of an XY Point with
	// X == 1.2 and Y == 2.3, converted between representations with
	// https://rodic.fr/blog/online-conversion-between-geometric-formats/
	testPointWKT     = []byte("POINT(1.2 2.3)")
	testPointGeoJSON = []byte(`{"type":"Point","coordinates":[1.2,2.3]}`)
	testPointWKB     = []byte{
		0x01, 0x01, 0x00, 0x00, 0x00, 0x33, 0x33, 0x33,
		0x33, 0x33, 0x33, 0xf3, 0x3f, 0x66, 0x66, 0x66,
		0x66, 0x66, 0x66, 0x02, 0x40,
	}
)

func TestSFPointCtors(t *testing.T) {
	require := require.New(t)

	// types.SFPoint is a wrapper around go-geom's Point class. As such,
	// construction typically uses their conventions.
	pa := types.NewSFPoint(
		*geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{1.2, 2.3}))
	require.Equal(1.2, pa.Lng())
	require.Equal(2.3, pa.Lat())
	require.Equal(
		*geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{1.2, 2.3}),
		pa.Point)

	// We have some helpers to make it easier, though.
	pb := types.NewSFPointXY(1.2, 2.3)
	require.Equal(
		*geom.NewPoint(geom.XY).MustSetCoords(geom.Coord{1.2, 2.3}),
		pb.Point)

	pc := types.NewSFPointXYZ(1.2, 2.3, 3.4)
	require.Equal(
		*geom.NewPoint(geom.XYZ).MustSetCoords(geom.Coord{1.2, 2.3, 3.4}),
		pc.Point)
}

func TestSFPointIsNil(t *testing.T) {
	require := require.New(t)

	p := types.NewSFPointXY(1.2, 2.3)
	require.False(p.IsNil())

	zero := types.NewSFPointXY(0.0, 0.0)
	require.False(zero.IsNil())

	empty := types.SFPoint{}
	require.True(empty.IsNil())
}

func TestSFPointIsZero(t *testing.T) {
	require := require.New(t)

	p := types.NewSFPointXY(1.2, 2.3)
	require.False(p.IsZero())

	zero := types.NewSFPointXY(0.0, 0.0)
	require.True(zero.IsZero())

	empty := types.SFPoint{}
	require.True(empty.IsZero())
}

func TestSFPointSQLValue(t *testing.T) {
	require := require.New(t)
	var val driver.Value
	var err error

	p := types.NewSFPointXY(1.2, 2.3)
	val, err = p.Value()
	require.NoError(err)
	require.EqualValues(testPointWKB, val)
}

func TestSFPointSQLScan(t *testing.T) {
	require := require.New(t)
	var err error

	var p types.SFPoint
	err = p.Scan(driver.Value(testPointWKB))
	require.NoError(err)
	require.Equal([]float64{1.2, 2.3}, p.FlatCoords())

	var bad types.SFPoint
	err = bad.Scan(driver.Value(nil))
	require.Error(err)
}

func TestSFPointMarshalJSON(t *testing.T) {
	require := require.New(t)
	var data []byte
	var err error

	p := types.NewSFPointXY(1.2, 2.3)
	data, err = json.Marshal(p)
	require.NoError(err)
	require.EqualValues(testPointGeoJSON, data)
	data, err = json.Marshal(&p)
	require.NoError(err)
	require.EqualValues(testPointGeoJSON, data)

	bad := types.SFPoint{}
	_, err = json.Marshal(bad)
	require.Error(err)
	_, err = json.Marshal(&bad)
	require.Error(err)
}

func TestSFPointUnmarshalJSON(t *testing.T) {
	require := require.New(t)
	var err error

	var p types.SFPoint
	err = json.Unmarshal(testPointGeoJSON, &p)
	require.NoError(err)
	require.Equal(1.2, p.Lng())
	require.Equal(2.3, p.Lat())
}

func TestSFPointMarshsalMapValue(t *testing.T) {
	require := require.New(t)
	type Wrapper struct{ Point types.SFPoint }
	var wrapper Wrapper
	var data map[string]interface{}
	var err error

	wrapper = Wrapper{types.NewSFPointXY(1.2, 2.3)}
	data, err = maps.Marshal(wrapper)
	require.NoError(err)
	require.Equal(types.NewSFPointXY(1.2, 2.3), data["Point"])
	data, err = maps.Marshal(&wrapper)
	require.NoError(err)
	require.Equal(types.NewSFPointXY(1.2, 2.3), data["Point"])
}
