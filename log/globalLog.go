package log

var glbLogger *Logger

func init() {
	l := NewLogger("", 3, "", "")
	SetGlobalLogger(l)
}

func SetGlobalLogger(l *Logger) {
	if glbLogger != nil {
		glbLogger.Close()
	}
	glbLogger = l
	glbLogger.callDepth += 1
}

// must be called after SetGlobalLogger()
func Warning(format string, args ...any) {
	glbLogger.Warning(format, args...)
}

func Error(format string, args ...any) {
	glbLogger.Error(format, args...)
}

func Critical(format string, args ...any) {
	glbLogger.Critical(format, args...)
}

func Evaluate(format string, args ...any) {
	glbLogger.Evaluate(format, args...)
}

func Info(format string, args ...any) {
	glbLogger.Info(format, args...)
}

func Stats(format string, args ...any) {
	glbLogger.Stats(format, args...)
}

func Audit(format string, args ...any) {
	glbLogger.Audit(format, args...)
}

func Transaction(format string, args ...any) {
	glbLogger.Transaction(format, args...)
}

func Debug(format string, args ...any) {
	glbLogger.Debug(format, args...)
}

func Trace(format string, args ...any) {
	glbLogger.Trace(format, args...)
}

func Verbose(format string, args ...any) {
	glbLogger.Verbose(format, args...)
}

func Close() {
	if glbLogger != nil {
		glbLogger.Close()
	}
}
