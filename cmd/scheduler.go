package cmd

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/orensimple/otus_events_scheduler/config"
	"github.com/orensimple/otus_events_scheduler/internal/grpc/api"
	"github.com/orensimple/otus_events_scheduler/internal/logger"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

var server, addr string

var RootCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Run scheduler events client",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(addr)
		logger.InitLogger()
		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			logger.ContextLogger.Errorf(" Cannot connect to GRPC server", err)
		}
		client := api.NewCalendarServiceClient(conn)
		req := &api.GetEventsByTimeRequest{
			TimeType: "month",
		}

		connAMQP, err := amqp.Dial("amqp://guest:guest@myapp-rabbitmq:5672/")
		if err != nil {
			logger.ContextLogger.Errorf("Failed to connect to RabbitMQ", err.Error())
		}
		defer connAMQP.Close()

		ch, err := connAMQP.Channel()
		if err != nil {
			logger.ContextLogger.Errorf("Failed to open a channel", err.Error())
		}
		defer ch.Close()

		err = ch.ExchangeDeclare(
			"eventsByDate", // name
			"direct",       // type
			true,           // durable
			false,          // auto-deleted
			false,          // internal
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			logger.ContextLogger.Infof("Failed to declare an exchange", err.Error())
		}

		_, err = ch.QueueDeclare(
			"eventsByDay", // name
			false,         // durable
			false,         // delete when unused
			false,         // exclusive
			false,         // no-wait
			nil,           // arguments
		)
		if err != nil {
			logger.ContextLogger.Infof("Problem QueueDeclare", err.Error())
		}

		err = ch.QueueBind(
			"eventsByDay",  // name
			"day",          // key
			"eventsByDate", // exchange
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			logger.ContextLogger.Infof("Problem bind queue", err.Error())
		}

		ctx := context.Background()
		logger.ContextLogger.Infof(" [*] Start to check events. To exit press CTRL+C")

		uptimeTicker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-uptimeTicker.C:
				resp, err := client.GetEventsByTime(context.Background(), req)
				if err != nil {
					log.Fatal(err)
				}
				if resp.GetError() != "" {
					log.Fatal(resp.GetError())
				} else {
					sendEvents(ctx, ch, resp)
				}
			}
		}

	},
}

func init() {
	RootCmd.Flags().StringVar(&addr, "config", "./config", "")
	RootCmd.Flags().StringVar(&server, "server", "events-api:8088", "host:port to connect to")
}

func sendEvents(ctx context.Context, ch *amqp.Channel, resp *api.GetEventsByTimeResponse) {

	for _, event := range resp.Event {
		err := ch.Publish(
			"eventsByDate", // exchange
			"day",          // routing key
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(strconv.FormatInt(event.GetID(), 10)),
			})

		if err != nil {
			logger.ContextLogger.Errorf("Failed to publish a message", err.Error())
			continue
		}
		logger.ContextLogger.Infof("Publish message", event.ID)
	}
}
