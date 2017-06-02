package wrouter

import (
	"reflect"
	"strings"
)

func controllerPath(controller interface{}) string {
	rc := reflect.TypeOf(controller)
	return strings.Replace(strings.ToLower(rc.Elem().Name()), "controller", "", -1)
}
