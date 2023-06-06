package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/consumerpool"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpscaling/cpuwork"
	"github.com/rabbitmq/amqp091-go"
	"github.com/struCoder/pidusage"
)

func main() {
	min := 1
	max := 20
	prefetch := 10

	var cpuUtilization int64

	go func() {
		t := time.NewTicker(time.Second)

		for range t.C {
			sysInfo, err := pidusage.GetStat(os.Getpid())
			if err != nil {
				log.Println("cpu usage:", err)
				continue
			}

			atomic.StoreInt64(&cpuUtilization, int64(sysInfo.CPU))
		}
	}()

	defDecider := consumerpool.DefaultDeciderFunc()

	cp, err := consumerpool.New(
		[]amqpextra.Option{
			amqpextra.WithURL("amqp://guest:guest@localhost:5672/test"),
		},
		[]consumer.Option{
			consumer.WithQos(prefetch, false),
			consumer.WithWorker(consumer.NewParallelWorker(prefetch)),
			consumer.WithDeclareQueue(`test-queue`, true, false, false, false, nil),
			consumer.WithHandler(consumer.HandlerFunc(func(ctx context.Context, msg amqp091.Delivery) interface{} {
				t := time.NewTimer(time.Second)
				defer t.Stop()

				tt := time.NewTicker(time.Millisecond * 10)
				defer tt.Stop()

				for {
					select {
					case <-tt.C:
						cpuwork.CPU(2)
					case <-t.C:
						if err := msg.Ack(false); err != nil {
							log.Printf("[ERROR] msg: ack: %s", err)
						}
						return nil
					}
				}
			})),
		},
		[]consumerpool.Option{
			consumerpool.WithLogger(logger.Std),
			consumerpool.WithMinSize(min),
			consumerpool.WithMaxSize(max),
			consumerpool.WithDecider(func(queueSize, poolSize, preFetch int) int {
				newPoolSize := defDecider(queueSize, poolSize, preFetch)

				cpuUtil := atomic.LoadInt64(&cpuUtilization)
				if cpuUtil > 60 {
					newPoolSize = poolSize - 1
				} else if cpuUtil > 50 {
					newPoolSize = poolSize
				}

				fmt.Println("cpu:", cpuUtil, "new:", newPoolSize, "curr:", poolSize)

				return newPoolSize
			}),
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer cp.Close()

	select {}
}
