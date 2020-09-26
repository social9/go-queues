package lib

import (
	"reflect"
)

// NonEmpty retuns the first non-empty value
//
// If no arguments are supplied, an interface with `0` value is returned
func NonEmpty(opts ...interface{}) interface{} {
	var ret interface{}
	ret = 0
	for _, opt := range opts {
		if !reflect.ValueOf(opt).IsZero() {
			ret = opt
			break
		}
	}
	return ret
}
