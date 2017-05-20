package wrouter

import (
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

func controllerPath(controller interface{}) string {
	rc := reflect.TypeOf(controller)
	return strings.Replace(strings.ToLower(rc.Elem().Name()), "controller", "", -1)
}
