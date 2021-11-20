package log

import "github.com/zuiwuchang/mget/widget"

func Layout() error {
	return defaultLog.Layout()
}
func Display() bool {
	return defaultLog.Display()
}
func Toggle(f func(string) *widget.Widget) {
	defaultLog.Toggle(f)
}
func Tagf(tag, format string, a ...interface{}) {
	defaultLog.Tagf(tag, format, a...)
}
func Errorf(format string, a ...interface{}) {
	defaultLog.Errorf(format, a...)
}
func Infof(format string, a ...interface{}) {
	defaultLog.Infof(format, a...)
}
func Tracef(format string, a ...interface{}) {
	defaultLog.Tracef(format, a...)
}

func Tag(tag string, a ...interface{}) {
	defaultLog.Tag(tag, a...)
}
func Error(a ...interface{}) {
	defaultLog.Error(a...)
}
func Info(a ...interface{}) {
	defaultLog.Info(a...)
}
func Trace(a ...interface{}) {
	defaultLog.Trace(a...)
}
