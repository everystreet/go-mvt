package geometry

import (
	"fmt"

	"github.com/everystreet/go-geojson"
)

// ToIntegers transforms a GeoJSON position to a pair of tile coordinates.
type ToIntegers func(geojson.Position) (x, y int32)

// Marshal returns the encoded sequence of a GeoJSON geometry.
func Marshal(v geojson.Geometry, transform ToIntegers) ([]uint32, error) {
	switch v := v.(type) {
	case *RawShape:
		return marshalRawShape(*v)
	case *geojson.Point:
		return marshalPoint(*v, transform)
	case *geojson.MultiPoint:
		return marshalMultiPoint(*v, transform)
	case *geojson.LineString:
		return marshalLineString(*v, transform)
	case *geojson.MultiLineString:
		return marshalMultiLineString(*v, transform)
	default:
		return nil, fmt.Errorf("unknown type '%t'", v)
	}
}

func marshalRawShape(v RawShape) ([]uint32, error) {
	return v, nil
}

func marshalPoint(v geojson.Point, transform ToIntegers) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, 1)
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(transform, geojson.Position(v))
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalMultiPoint(v geojson.MultiPoint, transform ToIntegers) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, uint32(len(v)))
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(transform, v...)
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalLineString(v geojson.LineString, transform ToIntegers) ([]uint32, error) {
	if len(v) < 2 {
		return nil, fmt.Errorf("linestring must consist of at least 2 points")
	}

	integers := make([]struct {
		x, y int32
	}, len(v))

	for i, pos := range v {
		integers[i].x, integers[i].y = transform(pos)
	}

	// MoveTo with command count == 1
	cmd, err := MakeCommandInteger(MoveTo, 1)
	if err != nil {
		return nil, err
	}
	data := []uint32{uint32(cmd)}

	// first point
	ints, err := marshalInteger(integers[0].x, integers[0].y)
	if err != nil {
		return nil, err
	}
	data = append(data, ints...)

	// LineTo with command count == remaining points
	cmd, err = MakeCommandInteger(LineTo, uint32(len(integers)-1))
	if err != nil {
		return nil, err
	}
	data = append(data, uint32(cmd))

	// remaining points
	for i := 1; i < len(integers); i++ {
		// points are relative to the previous point
		prev := integers[i-1]
		ints, err := marshalInteger(
			integers[i].x-prev.x,
			integers[i].y-prev.y)

		if err != nil {
			return nil, err
		}
		data = append(data, ints...)
	}

	return data, nil
}

func marshalMultiLineString(v geojson.MultiLineString, transform ToIntegers) ([]uint32, error) {
	var linestrings []uint32
	for _, line := range v {
		data, err := marshalLineString(line, transform)
		if err != nil {
			return nil, err
		}
		linestrings = append(linestrings, data...)
	}
	return linestrings, nil
}

func marshalPositions(transform ToIntegers, positions ...geojson.Position) ([]uint32, error) {
	var data []uint32
	for _, pos := range positions {
		x, y := transform(pos)

		integers, err := marshalInteger(x, y)
		if err != nil {
			return nil, err
		}
		data = append(data, integers...)
	}
	return data, nil
}

func marshalInteger(x, y int32) ([]uint32, error) {
	data := make([]uint32, 2)

	v, err := MakeParameterInteger(x)
	if err != nil {
		return nil, err
	}
	data[0] = uint32(v)

	v, err = MakeParameterInteger(y)
	if err != nil {
		return nil, err
	}
	data[1] = uint32(v)

	return data, nil
}
