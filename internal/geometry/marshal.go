package geometry

import (
	"fmt"

	"github.com/everystreet/go-geojson/v2"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s2"
)

// Project a geographic coordinate to a projected CRS.
type Project func(s2.LatLng) r2.Point

// Marshal returns the encoded sequence of a GeoJSON geometry.
func Marshal(v geojson.Geometry, project Project) ([]uint32, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}

	switch v := v.(type) {
	case *RawShape:
		return marshalRawShape(*v)
	case *geojson.Point:
		return marshalPoint(*v, project)
	case *geojson.MultiPoint:
		return marshalMultiPoint(*v, project)
	case *geojson.LineString:
		return marshalLineString(*v, project)
	case *geojson.MultiLineString:
		return marshalMultiLineString(*v, project)
	case *geojson.Polygon:
		return marshalPolygon(*v, project)
	case *geojson.MultiPolygon:
		return marshalMultiPolygon(*v, project)
	default:
		return nil, fmt.Errorf("unknown type '%t'", v)
	}
}

func marshalRawShape(v RawShape) ([]uint32, error) {
	return v, nil
}

func marshalPoint(v geojson.Point, project Project) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, 1)
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(project, geojson.Position(v))
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalMultiPoint(v geojson.MultiPoint, project Project) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, uint32(len(v)))
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(project, v...)
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalLineString(v geojson.LineString, project Project) ([]uint32, error) {
	if len(v) < 2 {
		return nil, fmt.Errorf("linestring must consist of at least 2 points")
	}

	points := make([]r2.Point, len(v))
	for i, p := range v {
		points[i] = project(p.LatLng)
	}

	// MoveTo with command count == 1
	cmd, err := MakeCommandInteger(MoveTo, 1)
	if err != nil {
		return nil, err
	}
	data := []uint32{uint32(cmd)}

	// first point
	ints, err := marshalInteger(points[0])
	if err != nil {
		return nil, err
	}
	data = append(data, ints...)

	// LineTo with command count == remaining points
	cmd, err = MakeCommandInteger(LineTo, uint32(len(points)-1))
	if err != nil {
		return nil, err
	}
	data = append(data, uint32(cmd))

	// remaining points
	for i := 1; i < len(points); i++ {
		// points are relative to the previous point
		prev := points[i-1]
		ints, err := marshalInteger(r2.Point{
			X: points[i].X - prev.X,
			Y: points[i].Y - prev.Y,
		})

		if err != nil {
			return nil, err
		}
		data = append(data, ints...)
	}

	return data, nil
}

func marshalMultiLineString(v geojson.MultiLineString, project Project) ([]uint32, error) {
	var linestrings []uint32
	for _, line := range v {
		data, err := marshalLineString(line, project)
		if err != nil {
			return nil, err
		}
		linestrings = append(linestrings, data...)
	}
	return linestrings, nil
}

func marshalPolygon(v geojson.Polygon, project Project) ([]uint32, error) {
	if len(v) < 1 {
		return nil, fmt.Errorf("polygon must consist of at least an exterior ring")
	}

	var data []uint32
	for i, loop := range v {
		// The first and last points of a GeoJSON polygon loop are the same,
		// but vector tiles implicitly connect the first and last points.
		// So we must remove last point and check we still have at least 3.
		if len(loop) < 4 {
			return nil, fmt.Errorf("loop '%d' must consist of at least 4 points (excluding the last)", i)
		}

		// A polygon loop is a linestring with a trailing ClosePath command.
		linestring, err := marshalLineString(loop[:len(loop)-1], project)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal loop '%d': %w", i, err)
		}
		data = append(data, linestring...)

		cmd, err := MakeCommandInteger(ClosePath, 1)
		if err != nil {
			return nil, err
		}
		data = append(data, uint32(cmd))
	}
	return data, nil
}

func marshalMultiPolygon(v geojson.MultiPolygon, project Project) ([]uint32, error) {
	var data []uint32
	for i, polygon := range v {
		p, err := marshalPolygon(polygon, project)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal polygon '%d': %w", i, err)
		}
		data = append(data, p...)
	}
	return data, nil
}

func marshalPositions(project Project, positions ...geojson.Position) ([]uint32, error) {
	var data []uint32
	for _, pos := range positions {
		point := project(pos.LatLng)

		integers, err := marshalInteger(point)
		if err != nil {
			return nil, err
		}
		data = append(data, integers...)
	}
	return data, nil
}

func marshalInteger(point r2.Point) ([]uint32, error) {
	data := make([]uint32, 2)

	v, err := MakeParameterInteger(int32(point.X))
	if err != nil {
		return nil, err
	}
	data[0] = uint32(v)

	v, err = MakeParameterInteger(int32(point.Y))
	if err != nil {
		return nil, err
	}
	data[1] = uint32(v)

	return data, nil
}
