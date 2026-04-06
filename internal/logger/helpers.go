package logger

import (
	"io"
	"log"
	"os"
)

func buildLogWriters(category string) (io.Writer, io.Writer, io.Writer, error) {
	_ = category
	return os.Stdout, os.Stderr, os.Stderr, nil
}

func newStdLogger(writer io.Writer, prefix string) *log.Logger {
	return log.New(writer, prefix, log.Ldate|log.Ltime|log.Lshortfile)
}
