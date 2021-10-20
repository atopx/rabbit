package client

import (
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type Config struct {
	hosts    string
	exchange string
	queue    string
	bind     bool
	delay    time.Duration
	args     amqp.Table
}

func NewConfig(hosts, queue string, bind bool, args amqp.Table) Config {
	var cfg = Config{hosts: hosts, queue: queue, bind: bind, args: args}
	if bind {
		cfg.exchange = queue
	}
	return cfg
}

type Client struct {
	mutex   *sync.RWMutex
	connect *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	config  Config
	closed  bool
}

func (c *Client) Close() {
	_ = c.channel.Close()
	_ = c.connect.Close()
}

func (c *Client) Open() (err error) {
	if c.connect, err = amqp.Dial(c.config.hosts); err != nil {
		return err
	}
	if c.channel, err = c.connect.Channel(); err != nil {
		return err
	}
	if c.queue, err = c.channel.QueueDeclare(c.config.queue, true, false, false, false, c.config.args); err != nil {
		return err
	}
	if c.config.bind {
		c.config.exchange = c.queue.Name
		if err = c.channel.QueueBind(c.queue.Name, c.queue.Name, c.queue.Name, false, c.config.args); err != nil {
			return err
		}
	}
	return nil
}

func New(config Config, connect *amqp.Connection) (*Client, error) {
	var s = Client{connect: connect, config: config, mutex: new(sync.RWMutex)}

	if err := s.Open(); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) Publish(msg amqp.Publishing) error {
	return c.channel.Publish(c.config.exchange, c.config.queue, false, false, msg)
}
