package geometry

import (
	"fmt"
	"reflect"

	"github.com/everystreet/go-geojson/v2"
	spec "github.com/everystreet/go-mvt/internal/spec"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s2"
)

// Unproject a projected coordinate to a geographic CRS.
type Unproject func(r2.Point) s2.LatLng

// Unmarshal parses the encoded geometry sequence and stores the result in the value pointed to by v.
func Unmarshal(data []uint32, typ spec.Tile_GeomType, unproject Unproject, v interface{}) error {
	rv, err := indirect(v)
	if err != nil {
		return err
	}

	geo, err := unmarshal(data, typ, unproject)
	if err != nil {
		return err
	}

	if rv.Kind() == reflect.Interface {
		rv.Set(reflect.ValueOf(geo))
	} else {
		rv.Set(reflect.Indirect(reflect.ValueOf(geo)))
	}
	return nil
}

func indirect(v interface{}) (*reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("v must be a pointer")
	}

	i := reflect.Indirect(rv)
	return &i, nil
}

func unmarshal(data []uint32, typ spec.Tile_GeomType, unproject Unproject) (geojson.Geometry, error) {
	switch typ {
	case spec.Tile_UNKNOWN:
		return (*RawShape)(&data), nil
	case spec.Tile_POINT:
		return unmarshalPoints(data, unproject)
	case spec.Tile_LINESTRING:
		return unmarshalLinestrings(data, unproject)
	case spec.Tile_POLYGON:
		return unmarshalPolygons(data, unproject)
	default:
		return nil, fmt.Errorf("unknown geometry type '%v'", typ)
	}
}

func unmarshalPoints(data []uint32, unproject Unproject) (geojson.Geometry, error) {
	n := len(data)
	if n == 0 {
		return nil, fmt.Errorf("data len must be >= 1")
	}

	cmd := CommandInteger(data[0])
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("invalid command: %w", err)
	} else if id := cmd.ID(); id != MoveTo {
		return nil, fmt.Errorf("expecting MoveTo command, received '%v'", id)
	}

	count := cmd.Count()
	switch {
	case count == 1 && n == 3:
		p, err := unmarshalPosition(data[1:], unproject)
		if err != nil {
			return nil, err
		}
		return (*geojson.Point)(p), nil
	case count > 1 && n == 1+int(count)*2:
		p, err := unmarshalPositions(data[1:], unproject)
		if err != nil {
			return nil, err
		}
		return (*geojson.MultiPoint)(&p), nil
	default:
		return nil, fmt.Errorf("MoveTo must be followed by at least one pair of ParameterIntegers: %d, %d", count, n)
	}
}

func unmarshalLinestrings(data []uint32, unproject Unproject) (geojson.Geometry, error) {
	var linestrings geojson.MultiLineString

	for len(data) != 0 {
		ls, err := unmarshalLineString(&data, unproject)
		if err != nil {
			return nil, err
		}
		linestrings = append(linestrings, *ls)
	}

	if len(linestrings) == 1 {
		return (*geojson.LineString)(&linestrings[0]), nil
	}
	return (*geojson.MultiLineString)(&linestrings), nil
}

func unmarshalPolygons(data []uint32, unproject Unproject) (geojson.Geometry, error) {
	var polygons geojson.MultiPolygon

	for len(data) != 0 {
		// A polygon loop is a linestring with a trailing ClosePath command.
		loop, err := unmarshalLineString(&data, unproject)
		if err != nil {
			return nil, err
		}

		if len(data) < 1 {
			return nil, fmt.Errorf("unexpected end")
		}

		// Consume the ClosePath.
		_, err = unmarshalCommand(data[0], ClosePath)
		if err != nil {
			return nil, err
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
				return nil, fmt.Errorf("missing exterior loop (%d)", len(*loop))
			}
			polygon := &polygons[len(polygons)-1]
			*polygon = append(*polygon, *loop)
		}
	}

	if len(polygons) == 1 {
		return (*geojson.Polygon)(&polygons[0]), nil
	}
	return (*geojson.MultiPolygon)(&polygons), nil
}

func unmarshalLineString(data *[]uint32, unproject Unproject) (*geojson.LineString, error) {
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

	points := make([]r2.Point, cmd.Count()+1)
	points[0] = r2.Point{
		X: float64(x.Value()),
		Y: float64(y.Value()),
	}

	// remaining coordinates make up the rest of the line
	for i := uint32(0); i < cmd.Count(); i++ {
		x, y, err := unmarshalIntegers((*data)[4+2*i : 6+2*i])
		if err != nil {
			return nil, err
		}

		// each coordinate is relative to the previous
		prev := points[i]
		points[i+1].X = prev.X + float64(x.Value())
		points[i+1].Y = prev.Y + float64(y.Value())
	}

	linestring := make(geojson.LineString, len(points))
	for i, p := range points {
		linestring[i] = geojson.Position{
			LatLng: unproject(p),
		}
	}

	*data = (*data)[lineDataLen:]
	return &linestring, nil
}

func unmarshalPositions(data []uint32, unproject Unproject) ([]geojson.Position, error) {
	if n := len(data); n%2 != 0 {
		return nil, fmt.Errorf("expecting even number of integers, have %d", n)
	}

	positions := make([]geojson.Position, len(data)/2)
	for i := 0; i < len(positions); i++ {
		pos, err := unmarshalPosition(data[i*2:i*2+2], unproject)
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

func unmarshalPosition(data []uint32, unproject Unproject) (*geojson.Position, error) {
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

	return &geojson.Position{
		LatLng: unproject(r2.Point{
			X: float64(x.Value()),
			Y: float64(y.Value()),
		}),
	}, nil
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
