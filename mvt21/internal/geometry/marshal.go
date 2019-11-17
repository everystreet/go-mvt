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
	default:
		return nil, fmt.Errorf("unknown type '%t'", v)
	}
}

func marshalRawShape(v RawShape) ([]uint32, error) {
	return nil, nil // TODO
}

func marshalPoint(point geojson.Point, transform ToIntegers) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, 1)
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(transform, geojson.Position(point))
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalMultiPoint(points geojson.MultiPoint, transform ToIntegers) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, uint32(len(points)))
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(transform, points...)
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalPositions(ints ToIntegers, positions ...geojson.Position) ([]uint32, error) {
	data := make([]uint32, 2*len(positions))
	for i, pos := range positions {
		x, y := ints(pos)

		v, err := MakeParameterInteger(x)
		if err != nil {
			return nil, err
		}
		data[i*2] = uint32(v)

		v, err = MakeParameterInteger(y)
		if err != nil {
			return nil, err
		}
		data[i*2+1] = uint32(v)
	}
	return data, nil
}
