package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davemolk/recordAPI/internal/data"
	"github.com/davemolk/recordAPI/internal/jsonlog"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	var cfg config

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	err := godotenv.Load()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DSN"), "Postgres DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections") 
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections") 
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	
	flag.Parse()

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connection established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)
	
	srv := &http.Server{
		Addr: fmt.Sprintf("localhost:%d", cfg.port),
		Handler: app.routes(),
		ErrorLog: log.New(logger, "", 0),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env": cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}