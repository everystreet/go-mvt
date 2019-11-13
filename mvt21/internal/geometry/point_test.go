package geometry_test

import (
	"fmt"
	"testing"

	"github.com/everystreet/go-mvt/mvt21/internal/geometry"
	"github.com/everystreet/go-geojson"
	"github.com/stretchr/testify/require"
)

func TestPoint(t *testing.T) {
	feature := geojson.NewPoint(12, 34)
	data, err := geometry.MarshalPoints(feature.Geometry, func(p geojson.Position) (x, y int32) {
		return int32(p.Longitude), int32(p.Latitude)
	})
	require.NoError(t, err)

	var point geojson.Point
	err = geometry.UnmarshalPoints(data, &point, func(x, y int32) geojson.Position {
		return geojson.Position{
			Longitude: float64(x),
			Latitude:  float64(y),
		}
	})
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &point)
}

func TestMultiPoint(t *testing.T) {
	feature := geojson.NewMultiPoint(
		geojson.NewPosition(12, 34),
		geojson.NewPosition(56, 78))

	data, err := geometry.MarshalPoints(feature.Geometry, func(p geojson.Position) (x, y int32) {
		return int32(p.Longitude), int32(p.Latitude)
	})
	require.NoError(t, err)
	fmt.Println(data)

	var point geojson.MultiPoint
	err = geometry.UnmarshalPoints(data, &point, func(x, y int32) geojson.Position {
		return geojson.Position{
			Longitude: float64(x),
			Latitude:  float64(y),
		}
	})
	require.NoError(t, err)
	require.Equal(t, feature.Geometry, &point)
}
