package geometry_test

import (
	"fmt"
	"testing"

	"github.com/everystreet/go-geojson"
	"github.com/everystreet/go-mvt/mvt21/internal/geometry"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestPoint(t *testing.T) {
	feature := geojson.NewPoint(12, 34)
	data, err := geometry.Marshal(feature.Geometry, func(p geojson.Position) (x, y int32) {
		return int32(p.Longitude), int32(p.Latitude)
	})
	require.NoError(t, err)

	var point geojson.Point
	err = geometry.Unmarshal(data, spec.Tile_POINT, func(x, y int32) geojson.Position {
		return geojson.Position{
			Longitude: float64(x),
			Latitude:  float64(y),
		}
	}, &point)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &point)
}

func TestMultiPoint(t *testing.T) {
	feature := geojson.NewMultiPoint(
		geojson.NewPosition(12, 34),
		geojson.NewPosition(56, 78))

	data, err := geometry.Marshal(feature.Geometry, func(p geojson.Position) (x, y int32) {
		return int32(p.Longitude), int32(p.Latitude)
	})
	require.NoError(t, err)
	fmt.Println(data)

	var points geojson.MultiPoint
	err = geometry.Unmarshal(data, spec.Tile_POINT, func(x, y int32) geojson.Position {
		return geojson.Position{
			Longitude: float64(x),
			Latitude:  float64(y),
		}
	}, &points)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &points)
}

func TestLineString(t *testing.T) {
	feature := geojson.NewLineString(
		geojson.NewPosition(12, 34),
		geojson.NewPosition(56, 78),
		geojson.NewPosition(90, 12),
		geojson.NewPosition(34, 56))
	data, err := geometry.Marshal(feature.Geometry, func(p geojson.Position) (x, y int32) {
		return int32(p.Longitude), int32(p.Latitude)
	})
	require.NoError(t, err)

	var linestring geojson.LineString
	err = geometry.Unmarshal(data, spec.Tile_LINESTRING, func(x, y int32) geojson.Position {
		return geojson.Position{
			Longitude: float64(x),
			Latitude:  float64(y),
		}
	}, &linestring)
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &linestring)
}
