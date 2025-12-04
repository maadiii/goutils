package log

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

func ZapExample() {
	cfg := Config{
		Level: DebugLevel,
		Writer: WriterConfig{
			Stdout: true,
			FileConfig: &lumberjack.Logger{
				Filename:   "app.log",
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     30,
				Compress:   true,
			},
		},

		QueueSize: 1000000,
		BatchSize: 1000,
		BatchDur:  time.Second,
	}

	log := Zap(cfg)

	start := time.Now()
	for i := 0; i < 1000000; i++ {
		log.Debug(context.TODO(), "Debug message", "user_id", i)
	}
	_ = log.Sync()

	fmt.Println("Elapsed:", time.Since(start))
}
