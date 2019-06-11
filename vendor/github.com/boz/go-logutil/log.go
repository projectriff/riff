package logutil

type Log interface {
	WithComponent(string) Log

	Trace(string, ...interface{}) string
	Un(string)

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})

	ErrWarn(error, string, ...interface{}) error
	ErrFatal(error, string, ...interface{}) error
	Err(error, string, ...interface{}) error
}
