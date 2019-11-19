package geometry_test

import (
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

	var raw geojson.Geometry
	err = geometry.Unmarshal(data, spec.Tile_UNKNOWN, SimpleFromIntegers, &raw)
	require.NoError(t, err)
	require.Equal(t, (*geometry.RawShape)(&data), raw.(*geometry.RawShape))

	data, err = geometry.Marshal(raw, SimpleToIntegers)
	require.NoError(t, err)
	require.Equal(t, raw, (*geometry.RawShape)(&data))
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

func TestPolygon(t *testing.T) {
	feature := geojson.NewPolygon(
		[]geojson.Position{
			geojson.NewPosition(7, 7),
			geojson.NewPosition(8, 4),
			geojson.NewPosition(4, 3),
			geojson.NewPosition(2, 5),
			geojson.NewPosition(3, 7),
			geojson.NewPosition(7, 7),
		},
		[]geojson.Position{
			geojson.NewPosition(4, 4),
			geojson.NewPosition(6, 4),
			geojson.NewPosition(7, 5),
			geojson.NewPosition(4, 6),
			geojson.NewPosition(4, 4),
		},
	)
	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)

	var polygon geojson.Polygon
	err = geometry.Unmarshal(data, spec.Tile_POLYGON, SimpleFromIntegers, &polygon)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &polygon)
}

func TestMultiPolygon(t *testing.T) {
	feature := geojson.NewMultiPolygon(
		[][]geojson.Position{
			{
				geojson.NewPosition(7, 7),
				geojson.NewPosition(8, 4),
				geojson.NewPosition(4, 3),
				geojson.NewPosition(2, 5),
				geojson.NewPosition(3, 7),
				geojson.NewPosition(7, 7),
			},
			[]geojson.Position{
				geojson.NewPosition(4, 4),
				geojson.NewPosition(6, 4),
				geojson.NewPosition(7, 5),
				geojson.NewPosition(4, 6),
				geojson.NewPosition(4, 4),
			},
		},
		[][]geojson.Position{
			{
				geojson.NewPosition(7, 7),
				geojson.NewPosition(4, 3),
				geojson.NewPosition(2, 5),
				geojson.NewPosition(7, 7),
			},
		},
	)
	data, err := geometry.Marshal(feature.Geometry, SimpleToIntegers)
	require.NoError(t, err)

	var multipolygon geojson.MultiPolygon
	err = geometry.Unmarshal(data, spec.Tile_POLYGON, SimpleFromIntegers, &multipolygon)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &multipolygon)
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
