package log

type Logger interface{
	Info(msg string)

	Warn(msg string)

	Error(msg string)

	InfoFormat(format string, vals ...interface{})

	WarnFormat(format string, vals ...interface{})

	ErrorFormat(format string, vals ...interface{})
}

var DebuggerLogger = &ConsoleLogger{}
var MuterLogger = &MuteLogger{}
var Factory = &LoggerFactory{}

type LoggerFactory struct {

}

func (lm *LoggerFactory) GetLogger() Logger{
	return DebuggerLogger
}
