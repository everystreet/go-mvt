package geometry

import (
	"fmt"
	"math"
)

// CommandID represents the command to be executed.
type CommandID uint8

const (
	// MoveTo creates a new point in a point geometry,
	// or starts a new vertex in a linestring or polygon geometry.
	MoveTo CommandID = 1
	// LineTo extends the current line or ring in a linestring or polygon geometry.
	LineTo CommandID = 2
	// ClosePath closes the current rin in a polygon geometry.
	ClosePath CommandID = 7
)

func (id CommandID) String() string {
	switch id {
	case MoveTo:
		return "MoveTo"
	case LineTo:
		return "LineTo"
	case ClosePath:
		return "ClosePath"
	default:
		return "unknown"
	}
}

// CommandInteger consists of a command ID, and the number of times to execute that command.
type CommandInteger uint32

// MakeCommandInteger encodes a CommandInteger from a command ID and count.
func MakeCommandInteger(id CommandID, count uint32) (CommandInteger, error) {
	if err := validateCommandInteger(id, count); err != nil {
		return 0, err
	}
	return CommandInteger((uint32(id) & 0x7) | (count << 3)), nil
}

// ID returns the encoded command ID.
func (i CommandInteger) ID() CommandID {
	return CommandID(i & 0x7)
}

// Count returns the encoded count.
func (i CommandInteger) Count() uint32 {
	return uint32(i >> 3)
}

// Validate the encoded command integer.
func (i CommandInteger) Validate() error {
	return validateCommandInteger(i.ID(), i.Count())
}

func validateCommandInteger(id CommandID, count uint32) error {
	switch id {
	case MoveTo, LineTo, ClosePath:
	default:
		return fmt.Errorf("invalid command ID, '%d'", id)
	}

	if max := uint32(math.Pow(2, 29) - 1); count > max {
		return fmt.Errorf("count exceeds maximum (%d > %d)", count, max)
	}
	return nil
}
