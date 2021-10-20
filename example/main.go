package main

import (
	"errors"
	"github.com/streadway/amqp"
	"github.com/yanmengfei/rabbit/client"
	"time"
)

func main() {
	config := client.NewConfig("amqp://admin:admin@127.0.0.1:5672/test", "worker", false, nil)
	cli, err := client.New(config, nil)
	if err != nil {
		panic(err)
	}
	for {
		if !errors.Is(cli.Publish(amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(`{"foo": "bar"}`),
		}), amqp.ErrClosed) {
			break
		}
		time.Sleep(3 * time.Second) // retry
	}
	cli.Close()
}
