package cmd

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/orensimple/otus_events_scheduler/config"
	"github.com/orensimple/otus_events_scheduler/internal/grpc/api"
	"github.com/orensimple/otus_events_scheduler/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

		// Create a metrics registry.
		reg := prometheus.NewRegistry()
		// Create some standard client metrics.
		grpcMetrics := grpc_prometheus.NewClientMetrics()
		// Register client metrics to registry.
		reg.MustRegister(grpcMetrics)

		conn, err := grpc.Dial(
			server,
			grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
			grpc.WithInsecure(),
		)
		if err != nil {
			logger.ContextLogger.Errorf("Cannot connect to GRPC server, retry after 30 second", err)
			timer1 := time.NewTimer(30 * time.Second)
			<-timer1.C
			conn, err = grpc.Dial(
				server,
				grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
				grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
				grpc.WithInsecure(),
			)
			if err != nil {
				logger.ContextLogger.Errorf("Cannot retry connect to GRPC server", err)
				os.Exit(1)
			}
		}
		client := api.NewCalendarServiceClient(conn)
		req := &api.GetEventsByTimeRequest{
			TimeType: "month",
		}

		connAMQP, err := amqp.Dial("amqp://guest:guest@myapp-rabbitmq:5672/")
		if err != nil {
			logger.ContextLogger.Errorf("Failed to connect to RabbitMQ, retry after 30 second", err.Error())
			timer1 := time.NewTimer(30 * time.Second)
			<-timer1.C
			connAMQP, err = amqp.Dial("amqp://guest:guest@myapp-rabbitmq:5672/")
			if err != nil {
				logger.ContextLogger.Errorf("Failed to retry connect to RabbitMQ", err.Error())
				os.Exit(1)
			}
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

		// Create a HTTP server for prometheus.
		httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: "events-scheduler:9120"}

		logger.ContextLogger.Infof("Starting web server at %s\n", "events-scheduler:9120")

		logger.ContextLogger.Infof(" [*] Start to check events. To exit press CTRL+C")
		go func() {
			uptimeTicker := time.NewTicker(5 * time.Second)
			for {
				select {
				case <-uptimeTicker.C:
					resp, err := client.GetEventsByTime(context.Background(), req)
					if err != nil {
						logger.ContextLogger.Errorf("Failed GetEventsByTime", err.Error())
					}
					if resp.GetError() != "" {
						logger.ContextLogger.Errorf("Failed response GetEventsByTime", resp.GetError())
					} else {
						sendEvents(ctx, ch, resp)
					}
				}
			}
		}()

		/*go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				logger.ContextLogger.Errorf("Unable to start a http server with metrics", err.Error())
			}
		}()*/
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
