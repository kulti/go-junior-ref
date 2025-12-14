package apiserver

import (
	"fmt"
	"os"
	"time"

	"github.com/kulti/task_list_course/internal/amqp"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/app"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/httpserver"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/notifier"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/store"
	"github.com/kulti/task_list_course/internal/pgconn"
	"github.com/kulti/task_list_course/internal/service"
)

func New() (*service.Service, error) {
	s := service.New()

	pgconn, err := pgconn.New(pgconn.Params{
		ConnectionString: os.Getenv("API_SERVER_DB_URL"),
		ChecknInterval:   5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("new db conn: %w", err)
	}
	s.AddComponent("pgconn", pgconn)

	publisher := amqp.NewPublisher(amqp.PublisherParams{
		DialAddress: os.Getenv("API_SERVER_AMQP_URL"),
		DeclareExchange: amqp.Exchange{
			Name: "item_events",
			Type: amqp.ExchangeFanout,
		},
	})
	s.AddComponent("publisher", publisher)

	consumer := amqp.NewConsumer(amqp.ConsumerParams{
		DialAddress: os.Getenv("API_SERVER_AMQP_URL"),
		DeclareExchange: amqp.Exchange{
			Name: "item_events",
			Type: amqp.ExchangeFanout,
		},
		DeclareQueue: amqp.Queue{
			Name:    "item_events.v1",
			Durable: true,
			BindTo: amqp.Bind{
				ExchangeName: "item_events",
			},
		},
	})
	s.AddComponent("consumer", consumer)

	store := store.New(store.Params{PgConn: pgconn})
	app := app.New(app.Params{Store: store, Publisher: publisher})

	server := httpserver.New(httpserver.Params{
		App:           app,
		ListenAddress: ":8090",
		ReadyHandler:  s.HandleReady,
	})
	s.AddComponent("httpserver", server)

	notifier := notifier.New(notifier.Params{Store: store, Consumer: consumer})
	s.AddComponent("notifier", notifier)

	return s, nil
}
