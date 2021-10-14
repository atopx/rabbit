package client

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"sync"
	"time"
)

var connect *amqp.Connection

type consumer struct {
	callback func(delivery amqp.Delivery)
	quit     chan bool
}

type Client struct {
	mutex    *sync.RWMutex
	url      string
	exchange string
	autoack  bool
	args     amqp.Table
	channel  *amqp.Channel
	config   *amqp.Config
	queue    amqp.Queue
	logger   *zap.Logger
	Delivery <-chan amqp.Delivery
	consumer consumer
}

func (c *Client) listenNotify() {
	delay := 5 * time.Second
	connectNotifyClose := connect.NotifyClose(make(chan *amqp.Error, 1))
	channelNotifyClose := c.channel.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case err := <-channelNotifyClose:
			if err != nil && err.Server {
				c.logger.Error("[channel]attempting to reconnect to amqp server after close", zap.Error(err))
				c.reconnect(delay)
			}

		case err := <-connectNotifyClose:
			if err != nil && err.Server {
				c.logger.Error("[connect]attempting to reconnect to amqp server after close", zap.Error(err))
				c.reconnect(delay)
			}
		}
		for range channelNotifyClose {
		}
		for range connectNotifyClose {
		}
	}
}

func (c *Client) reconnect(delay time.Duration) {
	var err error
	for {
		c.logger.Debug("waiting to attempt to reconnect rabbit server")
		c.mutex.Lock()
		if err = c.connect(c.queue.Name, c.queue.Name == c.exchange); err == nil {
			if c.Delivery != nil {
				c.startConsumer()
			}
			err = nil
			c.logger.Debug("reconnecting to amqp server success")
			c.mutex.Unlock()
			break
		}
		c.logger.Error("reconnecting to amqp server error", zap.Error(err))
		c.mutex.Unlock()
		time.Sleep(delay)
	}
}

func (c *Client) connect(name string, bind bool) (err error) {
	connect, c.channel, err = open(c.url, c.config)
	if err != nil {
		return err
	}
	queue, err := c.channel.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		return err
	}
	c.queue = queue
	if bind {
		c.exchange = name
		if err = c.channel.QueueBind(name, name, name, false, nil); err != nil {
			return err
		}
	}
	return nil
}

func open(url string, config *amqp.Config) (conn *amqp.Connection, ch *amqp.Channel, err error) {
	if conn == nil || !conn.IsClosed() {
		switch config {
		case nil:
			conn, err = amqp.Dial(url)
		default:
			conn, err = amqp.DialConfig(url, *config)
		}
	}
	if err != nil {
		return nil, nil, err
	}
	ch, err = conn.Channel()
	if err != nil {
		return nil, nil, err
	}
	return conn, ch, nil
}

func (c *Client) startConsumer() {
	c.Delivery, _ = c.channel.Consume(c.queue.Name, c.exchange, c.autoack, false, false, false, c.args)
	for delivery := range c.Delivery {
		c.consumer.callback(delivery)
	}
	c.logger.Debug("start consumer event")
	<-c.consumer.quit
	c.logger.Debug("stop consumer event")
}

func (c *Client) NewConsumer(callback func(delivery amqp.Delivery), autoack bool, args amqp.Table) {
	c.consumer = consumer{callback: callback, quit: make(chan bool, 1)}
	c.autoack = autoack
	c.args = args
	c.startConsumer()
}

func (c *Client) NewPublish(msg amqp.Publishing) error {
	return c.channel.Publish(c.exchange, c.queue.Name, false, false, msg)
}

func (c *Client) Clear(nowait bool) error {
	_, err := c.channel.QueuePurge(c.queue.Name, nowait)
	return err
}

func (c *Client) Close() {
	if c.consumer.quit != nil {
		c.consumer.quit <- true
	}
	_ = c.channel.Close()
	_ = connect.Close()
}

func New(url, queue string, bind bool, config *amqp.Config, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	c := Client{
		mutex:  &sync.RWMutex{},
		logger: logger,
		url:    url,
		config: config,
	}
	if err := c.connect(queue, bind); err != nil {
		return nil, err
	}
	go c.listenNotify()
	return &c, nil
}
