package debug

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func Println(v ...interface{}) {
	stacks()
	defaultLog.WriteString(stacks() +
		time.Now().Format(`2006/01/02 15:04:05 `) +
		fmt.Sprintln(v...),
	)
}
func Printf(format string, v ...interface{}) {
	stacks()
	defaultLog.WriteString(stacks() +
		time.Now().Format(`2006/01/02 15:04:05 `) +
		fmt.Sprintf(format, v...),
	)
}
func Print(v ...interface{}) {
	stacks()
	defaultLog.WriteString(stacks() +
		time.Now().Format(`2006/01/02 15:04:05 `) +
		fmt.Sprint(v...),
	)
}
func stacks() string {
	var strs []string
	for skip := 2; ; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		strs = append(strs, fmt.Sprintf(`- %v %v`,
			file, line,
		))
	}
	if len(strs) == 0 {
		return ``
	}
	return strings.Join(strs, "\n") + "\n"
}
