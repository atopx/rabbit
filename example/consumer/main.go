package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/yanmengfei/rabbit/client"
	"go.uber.org/zap"
)

const (
	url   = "amqp://admin:admin@127.0.0.1:5672/notice"
	queue = "notice"
)

var logger = zap.NewNop()

// 定义consumer handler
func handler(delivery amqp.Delivery) {
	fmt.Println(delivery.DeliveryMode)
	fmt.Println(string(delivery.Body))
}

func main() {
	// bind=true 兼容celery
	c, err := client.New(url, queue, false, nil, logger)
	if err != nil {
		logger.Panic(err.Error())
	}
	forever := make(chan bool)
	c.NewConsumer(handler, true, nil)
	<-forever
}
