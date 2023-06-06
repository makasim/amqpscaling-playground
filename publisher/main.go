package main

import (
	"flag"
	"log"
	"time"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	var rate int
	flag.IntVar(&rate, "rate", 10, "")
	flag.Parse()

	d, err := amqpextra.NewDialer(
		amqpextra.WithLogger(logger.Std),
		amqpextra.WithURL("amqp://guest:guest@localhost:5672/test"),
	)
	if err != nil {
		log.Fatal(err)
	}

	p, err := d.Publisher()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < rate; i++ {
		go func() {
			t := time.NewTicker(time.Second)
			for {
				start := time.Now()

				if err := p.Publish(publisher.Message{
					Key:          "test-queue",
					ErrOnUnready: false,
					Publishing: amqp091.Publishing{
						Body: []byte(`abody`),
					},
				}); err != nil {
					log.Fatal(err)
				}

				diff := time.Now().Sub(start)

				dur := time.Second - diff
				if dur <= 0 {
					dur = time.Second
				}
				t.Reset(dur)

				<-t.C
			}
		}()
		time.Sleep(time.Millisecond * 50)
	}

	select {}
}
