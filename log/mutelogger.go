package log

type MuteLogger struct {

}

func (c MuteLogger) Info(msg string){
}

func (c MuteLogger) Warn(msg string) {
}

func (c MuteLogger) Error(msg string) {
}


func (c MuteLogger) InfoFormat(format string, vals ...interface{}) {
}

func (c MuteLogger) WarnFormat(format string, vals ...interface{}) {
}

func (c MuteLogger) ErrorFormat(format string, vals ...interface{}) {
}