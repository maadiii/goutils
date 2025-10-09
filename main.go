package main

import (
	"context"
	"io"
	"os"

	"utils/log"
)

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	logger := log.New(&log.Config{
		Level:   log.DebugLevel,
		Writers: []io.Writer{file, os.Stdout},
	})

	logger.Error(context.Background(), "Hello World", "A", 1, "B", 2)
}
