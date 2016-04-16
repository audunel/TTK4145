package logger

import (
	"io"
	"log"
	"os"
)

func NewLogger(prefix string) log.Logger {
	fileFlag := os.O_RDWR | os.O_CREATE | os.O_APPEND
	file, err := os.OpenFile("log.txt", fileFlag, 0666)
	if err != nil {
		panic(err)
	}

	writer := io.MultiWriter(file, os.Stdout)

	return *log.New(writer, prefix+"\t", log.LstdFlags|log.Lshortfile)
}
