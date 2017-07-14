package log

import "fmt"

type ConsoleLogger struct {

}

func (c ConsoleLogger) Info(msg string){
	println(msg)
}

func (c ConsoleLogger) Warn(msg string) {
	println(msg)
}

func (c ConsoleLogger) Error(msg string) {
	println(msg)
}


func (c ConsoleLogger) InfoFormat(format string, vals ...interface{}) {
	fmt.Printf(format, vals...)
}

func (c ConsoleLogger) WarnFormat(format string, vals ...interface{}) {
	fmt.Printf(format, vals...)
}

func (c ConsoleLogger) ErrorFormat(format string, vals ...interface{}) {
	fmt.Printf(format, vals...)
}