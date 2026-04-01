package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewConsumer(url string, queueName string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		c.queue, // queue
		"",      // consumer tag
		false,   // auto ack
		false,   // exclusive
		false,   // no local
		false,   // no wait
		nil,     // args
	)
}

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
