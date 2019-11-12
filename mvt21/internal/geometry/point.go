package geometry

import (
	"fmt"
	"reflect"

	"github.com/mercatormaps/go-geojson"
)

// ToIntegers transforms a GeoJSON position to a pair of tile coordinates.
type ToIntegers func(geojson.Position) (x, y int32)

// MarshalPoints returns the encoded sequence of a GeoJSON point or multipoint.
func MarshalPoints(v geojson.Geometry, ints ToIntegers) ([]uint32, error) {
	switch v := v.(type) {
	case *geojson.Point:
		return marshalPoint(*v, ints)
	case *geojson.MultiPoint:
		return marshalMultiPoint(*v, ints)
	default:
		return nil, fmt.Errorf("unknown type '%T'", v)
	}
}

// FromIntegers transforms a pair of tile coordinates to a GeoJSON position.
type FromIntegers func(x, y int32) geojson.Position

// UnmarshalPoints parses the encoded point or mutlipoint and stores the result
// in the value pointed to by v.
func UnmarshalPoints(data []uint32, v geojson.Geometry, pos FromIntegers) error {
	rv, err := indirect(v)
	if err != nil {
		return err
	}

	if _, ok := v.(*RawShape); ok {
		rv.Set(reflect.ValueOf(data))
		return nil
	}

	n := len(data)
	if n == 0 {
		return fmt.Errorf("data len must be >= 1")
	}

	cmd := CommandInteger(data[0])
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	} else if id := cmd.ID(); id != MoveTo {
		return fmt.Errorf("expecting MoveTo command, received '%d'", id)
	}

	count := cmd.Count()
	switch {
	case count == 1 && n == 3:
		p, err := unmarshalPosition(data[1:], pos)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf((geojson.Point)(*p)))
	case count > 1 && n == 1+int(count)*2:
		p, err := unmarshalPositions(data[1:], pos)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf((geojson.MultiPoint)(p)))
	default:
		return fmt.Errorf("MoveTo must be followed by at least one pair of ParameterIntegers: %d, %d", count, n)
	}
	return nil
}

func marshalPoint(point geojson.Point, ints ToIntegers) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, 1)
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(ints, geojson.Position(point))
	if err != nil {
		return nil, err
	}
	return append([]uint32{uint32(cmd)}, positions...), nil
}

func marshalMultiPoint(points geojson.MultiPoint, ints ToIntegers) ([]uint32, error) {
	cmd, err := MakeCommandInteger(MoveTo, uint32(len(points)))
	if err != nil {
		return nil, err
	}

	positions, err := marshalPositions(ints, points...)
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

func unmarshalPositions(data []uint32, pos FromIntegers) ([]geojson.Position, error) {
	if n := len(data); n%2 != 0 {
		return nil, fmt.Errorf("expecting even number of integers, have %d", n)
	}

	positions := make([]geojson.Position, len(data)/2)
	for i := 0; i < len(positions); i++ {
		p, err := unmarshalPosition(data[i*2:i*2+2], pos)
		if err != nil {
			return nil, err
		}
		positions[i] = *p
	}
	return positions, nil
}

func unmarshalPosition(data []uint32, pos FromIntegers) (*geojson.Position, error) {
	if n := len(data); n != 2 {
		return nil, fmt.Errorf("expecting 2 integers, have %d", n)
	}

	x := ParameterInteger(data[0])
	if err := x.Validate(); err != nil {
		return nil, err
	}

	y := ParameterInteger(data[1])
	if err := y.Validate(); err != nil {
		return nil, err
	}

	p := pos(x.Value(), y.Value())
	return &p, nil
}

func indirect(v geojson.Geometry) (*reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr && rv.IsNil() {
		return nil, fmt.Errorf("v must be a pointer")
	}

	i := reflect.Indirect(rv)
	return &i, nil
}
