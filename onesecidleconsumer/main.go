package main

import (
	"context"
	"log"
	"time"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/consumerpool"
	"github.com/makasim/amqpextra/logger"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	min := 1
	max := 20
	prefetch := 10

	cp, err := consumerpool.New(
		[]amqpextra.Option{
			amqpextra.WithURL("amqp://guest:guest@localhost:5672/test"),
			amqpextra.WithLogger(logger.Std),
		},
		[]consumer.Option{
			consumer.WithLogger(logger.Std),
			consumer.WithQos(prefetch, false),
			consumer.WithWorker(consumer.NewParallelWorker(prefetch)),
			consumer.WithDeclareQueue(`test-queue`, true, false, false, false, nil),
			consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp091.Delivery) interface{} {
				time.Sleep(time.Second)
				if err := msg.Ack(false); err != nil {
					log.Printf("[ERROR] msg: ack: %s", err)
				}

				return nil
			})),
		},
		[]consumerpool.Option{
			consumerpool.WithLogger(logger.Std),
			consumerpool.WithMinSize(min),
			consumerpool.WithMaxSize(max),
			consumerpool.WithConsumersPerConn(3),
		},
	)
	if err != nil {
		log.Fatalf("consumerpool: new: %s", err)
	}
	defer cp.Close()

	select {}
}
