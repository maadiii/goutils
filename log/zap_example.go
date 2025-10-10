package log

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func ExampleAsync() {
	cfg := Config{
		Level: DebugLevel,
		Writers: []WriterConfig{
			{
				File: true,
				FileConfig: lumberjack.Logger{
					Filename:   "app.log",
					MaxSize:    100,
					MaxBackups: 3,
					MaxAge:     30,
					Compress:   true,
				},
			},
		},

		Async:     true,
		QueueSize: 100000,
		BatchSize: 1000,
		BatchDur:  time.Second,
	}

	log := Zap(cfg)

	start := time.Now()
	for i := 0; i < 1000000; i++ {
		log.Debug(context.TODO(), "Debug message", "user_id", i)
	}
	log.Sync()

	fmt.Println("Elapsed:", time.Since(start))
}
