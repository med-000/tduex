package logger

import "log"

type ScraperLogger struct {
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
}

func NewScraperLogger() (*ScraperLogger, error) {
	infoWriter, errorWriter, warnWriter, err := buildLogWriters("scraper")
	if err != nil {
		return nil, err
	}

	return &ScraperLogger{
		Info: newStdLogger(
			infoWriter,
			Blue+"[SCRAPER]"+Green+"[INFO] "+Reset,
		),
		Error: newStdLogger(
			errorWriter,
			Blue+"[SCRAPER]"+Red+"[ERROR] "+Reset,
		),
		Warn: newStdLogger(
			warnWriter,
			Blue+"[SCRAPER]"+Yellow+"[WARN] "+Reset,
		),
	}, nil
}
