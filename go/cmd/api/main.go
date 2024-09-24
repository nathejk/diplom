package main

import (
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"nathejk.dk/cmd/api/app"
	"nathejk.dk/internal/data"
	"nathejk.dk/internal/jsonlog"
	"nathejk.dk/internal/vcs"
	"nathejk.dk/nathejk/commands"
	"nathejk.dk/nathejk/table"
	"nathejk.dk/pkg/memorystream"
	"nathejk.dk/pkg/nats"
	"nathejk.dk/pkg/sqlpersister"
	"nathejk.dk/pkg/stream"
	"nathejk.dk/pkg/streaminterface"
)

var (
	version = vcs.Version()
)

// Define a config struct to hold all the configuration settings for our application.
type config struct {
	port    int
	webroot string
	db      struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	stan struct {
		dsn string
	}
}

type application struct {
	app.JsonApi

	config config
	models data.Models
	stan   streaminterface.Stream
	//publisher streaminterface.Publisher
	commands commands.Commands
	logger   *jsonlog.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 80, "API server port")
	flag.StringVar(&cfg.webroot, "webroot", getEnv("WEBROOT", "/app/public"), "Static web root")

	flag.StringVar(&cfg.stan.dsn, "stan-dsn", os.Getenv("STAN_DSN"), "NATS Streaming DSN")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "Database DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Database max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Database max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "Database max connection idle time")

	flag.Parse()

	//logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	/*
		js, err := jetstream.New(cfg.jetstream.dsn)
		if err != nil {
			log.Printf("Error connecting %q", err)
		}
		/*msg, err := js.LastMessage(streaminterface.SubjectFromStr("NATHEJK.2024.>"))
		if err != nil {
			log.Fatalf("Last message: %q", err)
		}
		log.Printf("Last message (%d) %v", msg.Sequence(), msg)
	*/
	natsstream := nats.NewNATSStreamUnique(cfg.stan.dsn, "diplom-api")
	defer natsstream.Close()

	db := NewDatabase(cfg.db)
	if err := db.Open(); err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	sqlw := sqlpersister.New(db.DB())

	memstream := memorystream.New()
	//bufferedPublisher := memstream
	dstmux := stream.NewStreamMux(memstream)
	dstmux.Handles(natsstream, "nathejk", "monolith") //d.stream.Channels()...)
	dstswtch, err := stream.NewSwitch(dstmux, []streaminterface.Consumer{
		table.NewPatrulje(sqlw),
		table.NewControlPoint(sqlw, memstream),
		table.NewControlGroupUser(sqlw, memstream),
		table.NewScan(sqlw, memstream),
	})
	//mux := xstream.NewMux(js)
	//mux.AddConsumer(table.NewSignup(sqlw), table.NewConfirm(sqlw), table.NewKlan(sqlw), table.NewSenior(sqlw), table.NewPatrulje(sqlw), table.NewPatruljeStatus(sqlw) /*table.NewPatruljeMerged(sqlw),*, table.NewSpejder(sqlw), table.NewSpejderStatus(sqlw))
	//mux.AddConsumer(table.NewSpejder(sqlw), table.NewSpejderStatus(sqlw))
	//if err := mux.Run(context.Background()); err != nil {
	//	logger.PrintFatal(err, nil)
	//}
	ctx := context.Background()
	live := make(chan struct{})
	go func() {
		err = dstswtch.Run(ctx, func() {
			//dstswtch.Close()
			//log.Printf("Closing")
			live <- struct{}{}
		})
		if err != nil {
			log.Fatal(err)
		}
	}()
	// Waiting for live
	select {
	case <-ctx.Done():
		log.Fatal(ctx.Err())
	case <-live:
	}

	models := data.NewModels(db.DB())

	expvar.NewString("version").Set(version)
	expvar.NewInt("timestamp").Set(time.Now().Unix())
	expvar.NewInt("goroutines").Set(int64(runtime.NumGoroutine()))

	app := &application{
		JsonApi: app.JsonApi{
			Logger: logger,
		},
		config: cfg,
		models: models,
		//jetstream: js,
		stan:     natsstream,
		commands: commands.New(natsstream, models),
		logger:   logger,
	}

	logger.PrintFatal(app.Serve(fmt.Sprintf(":%d", cfg.port), app.routes()), nil)
}
