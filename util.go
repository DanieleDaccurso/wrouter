package wrouter

import (
	"errors"
	"github.com/owtorg/events"
	"reflect"
	"strings"
)

// getHighestKey returns the highest key of a map which uses integers as keys. This is used
// to automatically append events to a router without having to specify the priority.
func getHighestKey(whatever map[int]events.Event) int {
	var highest = 0
	for k := range whatever {
		if k < highest {
			highest = k
		}
	}
	return highest
}

func verifyController(rc reflect.Type) error {
	if rc.Kind() != reflect.Ptr || rc.Elem().Kind() != reflect.Struct {
		return errors.New(rc.String() + " is a " + rc.Kind().String() + ", pointer to Struct expected")
	}

	if !strings.Contains(rc.String(), "Controller") {
		return errors.New(rc.String() + " does not end in Controller")
	}
	return nil
}

func controllerPath(controller interface{}) string {
	rc := reflect.TypeOf(controller)
	return strings.Replace(strings.ToLower(rc.Elem().Name()), "controller", "", -1)
}
