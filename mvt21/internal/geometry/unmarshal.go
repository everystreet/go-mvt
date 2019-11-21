package geometry

import (
	"fmt"
	"reflect"

	"github.com/everystreet/go-geojson/v2"
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
	case spec.Tile_LINESTRING:
		return unmarshalLinestrings(data, rv, transform)
	case spec.Tile_POLYGON:
		return unmarshalPolygons(data, rv, transform)
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
		return fmt.Errorf("expecting MoveTo command, received '%v'", id)
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

func unmarshalLinestrings(data []uint32, v *reflect.Value, transform FromIntegers) error {
	var linestrings geojson.MultiLineString

	for len(data) != 0 {
		ls, err := unmarshalLineString(&data, transform)
		if err != nil {
			return err
		}
		linestrings = append(linestrings, *ls)
	}

	if len(linestrings) == 1 {
		v.Set(reflect.ValueOf((geojson.LineString)(linestrings[0])))
	} else {
		v.Set(reflect.ValueOf((geojson.MultiLineString)(linestrings)))
	}
	return nil
}

func unmarshalPolygons(data []uint32, v *reflect.Value, transform FromIntegers) error {
	var polygons geojson.MultiPolygon

	for len(data) != 0 {
		// A polygon loop is a linestring with a trailing ClosePath command.
		loop, err := unmarshalLineString(&data, transform)
		if err != nil {
			return err
		}

		if len(data) < 1 {
			return fmt.Errorf("unexpected end")
		}

		// Consume the ClosePath.
		_, err = unmarshalCommand(data[0], ClosePath)
		if err != nil {
			return err
		}
		data = data[1:]

		//  GeoJSON loops are explicitly closed.
		if len(*loop) > 0 {
			*loop = append(*loop, (*loop)[0])
		}

		// Determine if this loop an exterior loop that starts a new polygon,
		// or an interior loop that belongs to the current polygon.
		if angle := geojson.LoopToS2(*loop).TurningAngle(); angle <= 0 { // CW exterior
			polygons = append(polygons, geojson.Polygon{*loop})
		} else if angle >= 0 { // CCW interior
			if len(polygons) == 0 {
				return fmt.Errorf("missing exterior loop")
			}
			polygon := &polygons[len(polygons)-1]
			*polygon = append(*polygon, *loop)
		}
	}

	if len(polygons) == 1 {
		v.Set(reflect.ValueOf((geojson.Polygon)(polygons[0])))
	} else {
		v.Set(reflect.ValueOf((geojson.MultiPolygon)(polygons)))
	}
	return nil
}

func unmarshalLineString(data *[]uint32, transform FromIntegers) (*geojson.LineString, error) {
	if n := len(*data); n < 4 {
		return nil, fmt.Errorf("data len must be >= 4, have %d", n)
	}

	// MoveTo with command count == 1
	cmd, err := unmarshalCommand((*data)[0], MoveTo)
	if err != nil {
		return nil, err
	} else if n := cmd.Count(); n != 1 {
		return nil, fmt.Errorf("expecting command count of 1, received '%d'", n)
	}

	// single pair for integers forms first coordinate
	x, y, err := unmarshalIntegers((*data)[1:3])
	if err != nil {
		return nil, err
	}

	// LineTo with command count >= 1
	cmd, err = unmarshalCommand((*data)[3], LineTo)
	if err != nil {
		return nil, err
	} else if n := cmd.Count(); n < 1 {
		return nil, fmt.Errorf("expecting command count >= 1, received '%d'", n)
	}

	// length of data for linestring
	lineDataLen := 4 + (2 * int(cmd.Count()))
	if n := len(*data); n < lineDataLen {
		return nil, fmt.Errorf("data len must be >= %d, have %d", lineDataLen, n)
	}

	points := make([]struct {
		x, y int32
	}, cmd.Count()+1)
	points[0].x = x.Value()
	points[0].y = y.Value()

	// remaining coordinates make up the rest of the line
	for i := uint32(0); i < cmd.Count(); i++ {
		x, y, err := unmarshalIntegers((*data)[4+2*i : 6+2*i])
		if err != nil {
			return nil, err
		}

		// each coordinate is relative to the previous
		prev := points[i]
		points[i+1].x = prev.x + x.Value()
		points[i+1].y = prev.y + y.Value()
	}

	linestring := make(geojson.LineString, len(points))
	for i, p := range points {
		linestring[i] = transform(p.x, p.y)
	}

	*data = (*data)[lineDataLen:]
	return &linestring, nil
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

func unmarshalIntegers(data []uint32) (x, y ParameterInteger, err error) {
	if n := len(data); n != 2 {
		return 0, 0, fmt.Errorf("expecting 2 integers, have %d", n)
	}

	x = ParameterInteger(data[0])
	if err := x.Validate(); err != nil {
		return 0, 0, err
	}

	y = ParameterInteger(data[1])
	if err := y.Validate(); err != nil {
		return 0, 0, err
	}
	return
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

func unmarshalCommand(data uint32, id CommandID) (*CommandInteger, error) {
	cmd := CommandInteger(data)
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command '%d': %w", data, err)
	} else if cmd.ID() != id {
		return nil, fmt.Errorf("expecting '%v' command, received '%v'", id, cmd.ID())
	}
	return &cmd, nil
}
