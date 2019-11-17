package geometry_test

import (
	"fmt"
	"testing"

	"github.com/everystreet/go-geojson"
	"github.com/everystreet/go-mvt/mvt21/internal/geometry"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestRawShape(t *testing.T) {
	feature := geojson.NewPoint(12, 34)
	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)

	var raw geometry.RawShape
	err = geometry.Unmarshal(data, spec.Tile_UNKNOWN, SimpleFromIntegers, &raw)
	require.NoError(t, err)
	require.Equal(t, data, []uint32(raw))

	data, err = geometry.Marshal(&raw, SimpleToIntegers)
	require.NoError(t, err)
	require.Equal(t, raw, geometry.RawShape(data))
}

func TestPoint(t *testing.T) {
	feature := geojson.NewPoint(12, 34)
	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)

	var point geojson.Point
	err = geometry.Unmarshal(data, spec.Tile_POINT, SimpleFromIntegers, &point)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &point)
}

func TestMultiPoint(t *testing.T) {
	feature := geojson.NewMultiPoint(
		geojson.NewPosition(12, 34),
		geojson.NewPosition(56, 78))

	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)
	fmt.Println(data)

	var points geojson.MultiPoint
	err = geometry.Unmarshal(data, spec.Tile_POINT, SimpleFromIntegers, &points)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &points)
}

func TestLineString(t *testing.T) {
	feature := geojson.NewLineString(
		geojson.NewPosition(12, 34),
		geojson.NewPosition(56, 78),
		geojson.NewPosition(90, 12),
		geojson.NewPosition(34, 56))
	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)

	var linestring geojson.LineString
	err = geometry.Unmarshal(data, spec.Tile_LINESTRING, SimpleFromIntegers, &linestring)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &linestring)
}

func TestMultiLineString(t *testing.T) {
	feature := geojson.NewMultiLineString(
		[]geojson.Position{
			geojson.NewPosition(12, 34),
			geojson.NewPosition(56, 78),
			geojson.NewPosition(90, 12),
		},
		[]geojson.Position{
			geojson.NewPosition(23, 45),
			geojson.NewPosition(67, 89),
			geojson.NewPosition(12, 34),
			geojson.NewPosition(56, 78),
		},
	)
	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)

	var multilinestring geojson.MultiLineString
	err = geometry.Unmarshal(data, spec.Tile_LINESTRING, SimpleFromIntegers, &multilinestring)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &multilinestring)
}

var SimpleToIntegers = func(pos geojson.Position) (x, y int32) {
	return int32(pos.Longitude) - 10, int32(pos.Latitude) - 10
}

var SimpleFromIntegers = func(x, y int32) geojson.Position {
	return geojson.Position{
		Longitude: float64(x) + 10,
		Latitude:  float64(y) + 10,
	}
}
