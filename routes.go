package wrouter

import "reflect"

// Route represents one callable route
type Route struct {
	// Methods contains a slice of strings, identifying the allowed HTTP methods for the current route
	Methods []string
	// Controller contains the current Controller
	Controller Controller
	// RMethod contains the reflect.Method of the Controller's action to be called on the current route
	RMethod reflect.Method
	// Path contains the path for the current route
	Path string
}
