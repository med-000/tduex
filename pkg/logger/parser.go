package logger

import "log"

type ParserLogger struct {
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
}

func NewParserLogger() (*ParserLogger, error) {
	infoWriter, errorWriter, warnWriter, err := buildLogWriters("parser")
	if err != nil {
		return nil, err
	}

	return &ParserLogger{
		Info: newStdLogger(
			infoWriter,
			Cyan+"[PARSER]"+Green+"[INFO] "+Reset,
		),
		Error: newStdLogger(
			errorWriter,
			Cyan+"[PARSER]"+Red+"[ERROR] "+Reset,
		),
		Warn: newStdLogger(
			warnWriter,
			Cyan+"[PARSER]"+Yellow+"[WARN] "+Reset,
		),
	}, nil
}
