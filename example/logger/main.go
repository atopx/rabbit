package main

import (
	"github.com/streadway/amqp"
	"github.com/yanmengfei/rabbit/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	url   = "amqp://admin:admin@127.0.0.1:5672/notice"
	queue = "notice"
)

var logger *zap.Logger

func SetupLogger(level string) (err error) {
	var loggerLevel = new(zapcore.Level)
	if err = loggerLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), loggerLevel)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return err
}

// 定义consumer handler
func handler(delivery amqp.Delivery) {
	logger.Info(string(delivery.Body))
}

func main() {
	_ = SetupLogger("debug")
	c, err := client.New(url, queue, false, nil, logger)
	if err != nil {
		logger.Panic(err.Error())
	}
	forever := make(chan bool)
	c.NewConsumer(handler, true, nil)
	<-forever
}
