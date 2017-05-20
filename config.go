package wrouter

import (
	"io"
	"os"
)

// Configuration contains the configurable settings of the router
type Configuration struct {
	// ErrorRedirect when set to true and using an error controller, will redirect
	// the error request to the standard error URL if an ErrorController is defined and its
	// default path has not been changed. If set to false, the ErrorController will simply be called under
	// the current path without doing an actual redirect.
	//
	// Default: false
	ErrorRedirect bool

	// CreateAliasRoutes when set to true, will create alias routes for the IndexController (being callable by /)
	// and IndexActions in every controller (being called by the controller path).
	// Disabling when alias-routes are not needed, will have minor performance improvements on build and runtime.
	//
	// Default: true
	CreateAliasRoutes bool

	// AllowSubController when set to true, will allow embedding a controller into another controller to create
	// multi-level routes. If set to false, you will only be able to generate routes on two levels by convention.
	// This does not influence the ability to change routes in general.
	// Disabling will have minor performance improvements on build, but none on runtime.
	//
	// Default: true
	AllowSubController bool

	// Verbosity contains all information about the verbosity of the Router. By default, Verbose settings are off.
	// Enabling verbosity, depending on your verbosity settings, may have measurable performance drawbacks.
	Verbosity struct {
		// Verbose when set to true, will make the router print various build & runtime information to the
		// configured io.Writer. See Configuration.Verbosity.Writer to set a writer.
		//
		// Default: false
		Verbose bool

		// SyncVerbose when set to false, will open a new goroutine for every verbose output instead of blocking
		// the application. Consider using this especially if your io.Writer is slow. Not syncing the log
		// might produce questionable output.
		//
		// Default: true
		SyncVerbose bool

		// Writer is the writer to which the verbose output it printed.
		//
		// Default: os.Stdout
		Writer io.Writer
	}
}

func createDefaultConfiguration() *Configuration {
	c := new(Configuration)
	c.ErrorRedirect = false
	c.CreateAliasRoutes = true
	c.AllowSubController = true
	c.Verbosity.SyncVerbose = false
	c.Verbosity.SyncVerbose = true
	c.Verbosity.Writer = os.Stdout
	return c
}
