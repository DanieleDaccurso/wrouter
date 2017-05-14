# Lifecycle context proposal

Make an object available to the entire lifecycle of the call, for passing through data.

Currently this implements 1 reflection heavy way of managing this to be used as a jumping off point for discussion.

Context is passed to all events and routes, this contains request, response writer, map[string]interface{} storage array and possible []reflect.Value results array.
This keeps the method signatures consistent between pre/post events and routes.

Current implementation is extremely low tech and not production ready in any way!

tom_test.go will run current tests, which simply test that it works at all without throwing an exception.

stringresponsewriter would need to be refactored to work with this method