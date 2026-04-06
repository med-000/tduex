package logger

import "log"

type ServiceLogger struct {
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
}

func NewServiceLogger() (*ServiceLogger, error) {
	infoWriter, errorWriter, warnWriter, err := buildLogWriters("Service")
	if err != nil {
		return nil, err
	}

	return &ServiceLogger{
		Info: newStdLogger(
			infoWriter,
			BrightBlue+"[SERVICE]"+Green+"[INFO] "+Reset,
		),
		Error: newStdLogger(
			errorWriter,
			BrightBlue+"[SERVICE]"+Red+"[ERROR] "+Reset,
		),
		Warn: newStdLogger(
			warnWriter,
			BrightBlue+"[SERVICE]"+Yellow+"[WARN] "+Reset,
		),
	}, nil
}
