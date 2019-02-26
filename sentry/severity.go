package sentry

// Severity stands for logging level
// http://docs.python.org/2/howto/logging.html#logging-levels
type Severity string

// Predefined levels
const (
	DEBUG   = Severity("debug")
	INFO    = Severity("info")
	WARNING = Severity("warning")
	ERROR   = Severity("error")
	FATAL   = Severity("fatal")
)
