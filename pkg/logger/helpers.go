package logger

import (
	"io"
	"log"
	"os"
)

func buildLogWriters(category string) (io.Writer, io.Writer, io.Writer, error) {
	infoFile, err := getRotatingWriter("logs/" + category + "/info.log")
	if err != nil {
		return nil, nil, nil, err
	}

	errorFile, err := getRotatingWriter("logs/" + category + "/error.log")
	if err != nil {
		return nil, nil, nil, err
	}

	warnFile, err := getRotatingWriter("logs/" + category + "/warn.log")
	if err != nil {
		return nil, nil, nil, err
	}

	appFile, err := getRotatingWriter("logs/app.log")
	if err != nil {
		return nil, nil, nil, err
	}

	infoWriter := io.MultiWriter(os.Stdout, infoFile, appFile)
	errorWriter := io.MultiWriter(os.Stderr, errorFile, appFile)
	warnWriter := io.MultiWriter(os.Stderr, warnFile, appFile)

	return infoWriter, errorWriter, warnWriter, nil
}

func newStdLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime|log.Lshortfile)
}
