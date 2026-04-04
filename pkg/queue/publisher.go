package queue

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewPublisher(url string, queueName string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto delete
		false,     // exclusive
		false,     // no wait
		nil,       // args
	)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Queue connected")

	return &Publisher{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, paymentID string) error {
	return p.channel.PublishWithContext(ctx,
		"",      // exchange
		p.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(paymentID),
			DeliveryMode: amqp.Persistent, // survive RabbitMQ restart
		},
	)
}

// ← new DLQ publish function
func (p *Publisher) PublishToDLQ(ctx context.Context, paymentID string) error {
	dlqName := p.queue + "_dead"

	// declare DLQ
	_, err := p.channel.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx,
		"",
		dlqName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(paymentID),
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (p *Publisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
