package rfm9x

import "log"

type Log_level uint

const (
	LogLevelRegIO Log_level = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelErr
)

// Based on the contents of https://stackoverflow.com/a/47515351/9541128.
type rfm9x_logger struct {
	level Log_level
}

func (l *rfm9x_logger) reg_io(fmt string, args ...interface{}) {
	if l.level <= LogLevelRegIO {
		log.Printf("* REG I/O * "+fmt, args...)
	}
}

func (l *rfm9x_logger) debug(fmt string, args ...interface{}) {
	if l.level < LogLevelInfo {
		log.Printf("* DEBUG * "+fmt, args...)
	}
}

func (l *rfm9x_logger) warn(fmt string, args ...interface{}) {
	log.Printf("* WARNING * "+fmt, args...)
}
