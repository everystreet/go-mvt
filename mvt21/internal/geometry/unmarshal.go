package geometry

import (
	"fmt"
	"reflect"

	"github.com/everystreet/go-geojson"
	spec "github.com/everystreet/go-mvt/mvt21/internal/spec"
)

// FromIntegers transforms a pair of tile coordinates to a GeoJSON position.
type FromIntegers func(x, y int32) geojson.Position

// Unmarshal parses the encoded geometry sequence and stores the result in the value pointed to by v.
func Unmarshal(data []uint32, typ spec.Tile_GeomType, transform FromIntegers, v interface{}) error {
	rv, err := indirect(v)
	if err != nil {
		return err
	}

	switch typ {
	case spec.Tile_UNKNOWN:
		return unmarshalRawShape(data, rv)
	case spec.Tile_POINT:
		return unmarshalPoints(data, rv, transform)
	default:
		return fmt.Errorf("unknown geometry type '%v'", typ)
	}
}

func indirect(v interface{}) (*reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("v must be a pointer")
	}

	i := reflect.Indirect(rv)
	return &i, nil
}

func unmarshalRawShape(data []uint32, v *reflect.Value) error {
	raw := RawShape(data)
	v.Set(reflect.ValueOf(&raw))
	return nil
}

func unmarshalPoints(data []uint32, v *reflect.Value, transform FromIntegers) error {
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
		p, err := unmarshalPosition(data[1:], transform)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf((geojson.Point)(*p)))
	case count > 1 && n == 1+int(count)*2:
		p, err := unmarshalPositions(data[1:], transform)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf((geojson.MultiPoint)(p)))
	default:
		return fmt.Errorf("MoveTo must be followed by at least one pair of ParameterIntegers: %d, %d", count, n)
	}
	return nil
}

func unmarshalPositions(data []uint32, transform FromIntegers) ([]geojson.Position, error) {
	if n := len(data); n%2 != 0 {
		return nil, fmt.Errorf("expecting even number of integers, have %d", n)
	}

	positions := make([]geojson.Position, len(data)/2)
	for i := 0; i < len(positions); i++ {
		pos, err := unmarshalPosition(data[i*2:i*2+2], transform)
		if err != nil {
			return nil, err
		}
		positions[i] = *pos
	}
	return positions, nil
}

func unmarshalPosition(data []uint32, transform FromIntegers) (*geojson.Position, error) {
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

	pos := transform(x.Value(), y.Value())
	return &pos, nil
}
