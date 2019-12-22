package geometry_test

import (
	"testing"

	"github.com/everystreet/go-geojson/v2"
	"github.com/everystreet/go-mvt/internal/geometry"
	spec "github.com/everystreet/go-mvt/internal/spec"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/require"
)

func TestRawShape(t *testing.T) {
	feature := geojson.NewPoint(34, 12)
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var raw geojson.Geometry
	err = geometry.Unmarshal(data, spec.Tile_UNKNOWN, SimpleUnproject, &raw)
	require.NoError(t, err)
	require.Equal(t, (*geometry.RawShape)(&data), raw.(*geometry.RawShape))

	data, err = geometry.Marshal(raw, SimpleProject)
	require.NoError(t, err)
	require.Equal(t, raw, (*geometry.RawShape)(&data))
}

func TestPoint(t *testing.T) {
	feature := geojson.NewPoint(34, 12)
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var point geojson.Point
	err = geometry.Unmarshal(data, spec.Tile_POINT, SimpleUnproject, &point)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &point)
}

func TestMultiPoint(t *testing.T) {
	feature := geojson.NewMultiPoint(
		geojson.MakePosition(34, 12),
		geojson.MakePosition(78, 56))

	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var points geojson.MultiPoint
	err = geometry.Unmarshal(data, spec.Tile_POINT, SimpleUnproject, &points)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &points)
}

func TestLineString(t *testing.T) {
	feature := geojson.NewLineString(
		geojson.MakePosition(34, 12),
		geojson.MakePosition(78, 56),
		geojson.MakePosition(12, 90),
		geojson.MakePosition(56, 34))
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var linestring geojson.LineString
	err = geometry.Unmarshal(data, spec.Tile_LINESTRING, SimpleUnproject, &linestring)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &linestring)
}

func TestMultiLineString(t *testing.T) {
	feature := geojson.NewMultiLineString(
		[]geojson.Position{
			geojson.MakePosition(12, 34),
			geojson.MakePosition(56, 78),
			geojson.MakePosition(90, 12),
		},
		[]geojson.Position{
			geojson.MakePosition(23, 45),
			geojson.MakePosition(67, 89),
			geojson.MakePosition(12, 34),
			geojson.MakePosition(56, 78),
		},
	)
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var multilinestring geojson.MultiLineString
	err = geometry.Unmarshal(data, spec.Tile_LINESTRING, SimpleUnproject, &multilinestring)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &multilinestring)
}

func TestPolygon(t *testing.T) {
	feature := geojson.NewPolygon(
		[]geojson.Position{
			geojson.MakePosition(7, 7),
			geojson.MakePosition(4, 8),
			geojson.MakePosition(3, 4),
			geojson.MakePosition(5, 2),
			geojson.MakePosition(7, 3),
			geojson.MakePosition(7, 7),
		},
		[]geojson.Position{
			geojson.MakePosition(4, 4),
			geojson.MakePosition(4, 6),
			geojson.MakePosition(5, 7),
			geojson.MakePosition(6, 4),
			geojson.MakePosition(4, 4),
		},
	)
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var polygon geojson.Polygon
	err = geometry.Unmarshal(data, spec.Tile_POLYGON, SimpleUnproject, &polygon)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &polygon)
}

func TestMultiPolygon(t *testing.T) {
	feature := geojson.NewMultiPolygon(
		[][]geojson.Position{
			{
				geojson.MakePosition(7, 7),
				geojson.MakePosition(4, 8),
				geojson.MakePosition(3, 4),
				geojson.MakePosition(5, 2),
				geojson.MakePosition(7, 3),
				geojson.MakePosition(7, 7),
			},
			[]geojson.Position{
				geojson.MakePosition(4, 4),
				geojson.MakePosition(4, 6),
				geojson.MakePosition(5, 7),
				geojson.MakePosition(6, 4),
				geojson.MakePosition(4, 4),
			},
		},
		[][]geojson.Position{
			{
				geojson.MakePosition(7, 7),
				geojson.MakePosition(3, 4),
				geojson.MakePosition(5, 2),
				geojson.MakePosition(7, 7),
			},
		},
	)
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var multipolygon geojson.MultiPolygon
	err = geometry.Unmarshal(data, spec.Tile_POLYGON, SimpleUnproject, &multipolygon)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &multipolygon)
}

func TestGeneric(t *testing.T) {
	feature := geojson.NewPoint(34, 12)
	data, err := geometry.Marshal(feature.Geometry, SimpleProject)
	require.NoError(t, err)

	var point geojson.Geometry
	err = geometry.Unmarshal(data, spec.Tile_POINT, SimpleUnproject, &point)
	require.NoError(t, err)
	require.IsType(t, &geojson.Point{}, point)
	require.Equal(t, feature.Geometry, point.(*geojson.Point))
}

var SimpleProject = func(ll s2.LatLng) r2.Point {
	return r2.Point{
		X: ll.Lng.Degrees() - 10,
		Y: ll.Lat.Degrees() - 10,
	}
}

var SimpleUnproject = func(p r2.Point) s2.LatLng {
	return s2.LatLngFromDegrees(p.Y+10, p.X+10)
}
