// Code generated by thriftrw v1.2.0
// @generated

package kv

import (
	"errors"
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

type ResourceDoesNotExist struct {
	Key     string  `json:"key"`
	Message *string `json:"message,omitempty"`
}

func (v *ResourceDoesNotExist) ToWire() (wire.Value, error) {
	var (
		fields [2]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)
	w, err = wire.NewValueString(v.Key), error(nil)
	if err != nil {
		return w, err
	}
	fields[i] = wire.Field{ID: 1, Value: w}
	i++
	if v.Message != nil {
		w, err = wire.NewValueString(*(v.Message)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 2, Value: w}
		i++
	}
	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func (v *ResourceDoesNotExist) FromWire(w wire.Value) error {
	var err error
	keyIsSet := false
	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				v.Key, err = field.Value.GetString(), error(nil)
				if err != nil {
					return err
				}
				keyIsSet = true
			}
		case 2:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Message = &x
				if err != nil {
					return err
				}
			}
		}
	}
	if !keyIsSet {
		return errors.New("field Key of ResourceDoesNotExist is required")
	}
	return nil
}

func (v *ResourceDoesNotExist) String() string {
	if v == nil {
		return "<nil>"
	}
	var fields [2]string
	i := 0
	fields[i] = fmt.Sprintf("Key: %v", v.Key)
	i++
	if v.Message != nil {
		fields[i] = fmt.Sprintf("Message: %v", *(v.Message))
		i++
	}
	return fmt.Sprintf("ResourceDoesNotExist{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {
		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func (v *ResourceDoesNotExist) Equals(rhs *ResourceDoesNotExist) bool {
	if !(v.Key == rhs.Key) {
		return false
	}
	if !_String_EqualsPtr(v.Message, rhs.Message) {
		return false
	}
	return true
}

func (v *ResourceDoesNotExist) Error() string {
	return v.String()
}
