package main

import (
	"github.com/streadway/amqp"
	"github.com/yanmengfei/rabbit/client"
)

const (
	url   = "amqp://admin:admin@127.0.0.1:5672/notice"
	queue = "notice"
)

func main() {
	// bind=true 兼容celery
	c, err := client.New(url, queue, false, nil, nil)
	if err != nil {
		panic(err)
	}
	if err = c.NewPublish(amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         []byte(`{"operate":"create","vulkey":"SOC-2021-11111","pockey":"POC-2021-11111"}`),
	}); err != nil {
		panic(err)
	}
}
