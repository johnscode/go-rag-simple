package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"log"
	"os"
	"strings"
	"time"
)

func setupLogger(ctx context.Context, logFilePath string) *zerolog.Logger {
	var outWriter = os.Stdout
	if logFilePath != "" && logFilePath != "stdout" {
		file, err := os.OpenFile(logFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatalln(err)
		}
		outWriter = file
	}
	cout := zerolog.ConsoleWriter{Out: outWriter, TimeFormat: time.RFC822}
	cout.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	// uncomment to remove timestamp from logs
	//out.FormatTimestamp = func(i interface{}) string {
	//	return ""
	//}
	baseLogger := zerolog.New(cout).With().Timestamp().Logger()
	logCtx := baseLogger.WithContext(ctx)
	l := zerolog.Ctx(logCtx)
	return l
}
