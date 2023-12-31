package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/maxwellkuo47/tradingEngine/internal/data"
	"github.com/redis/go-redis/v9"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	consumer struct {
		frequncy uint
	}
}

type application struct {
	config          config
	errorLogger     *slog.Logger
	infoLogger      *slog.Logger
	models          data.DBModels
	wg              sync.WaitGroup
	redisClient     *redis.Client
	mockStockPrices sync.Map
	done            chan bool
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// db config
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://trading:pa55w0rd@localhost/trading?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 200, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 200, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// rate lmiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 6, "Rate limiter maximum request per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 10, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")

	// consumer frequency
	flag.UintVar(&cfg.consumer.frequncy, "consumer-frequncy", 50, "Consumer frequency")

	// parsing flag
	flag.Parse()
	infoLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	errorLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelError,
		AddSource: true,
	}))

	db, err := OpenDB(cfg)
	if err != nil {
		errorLogger.Error("OpenDB error", slog.String("msg", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	infoLogger.Info("DB Connection", slog.String("Status", "OK"))

	redis, err := createRedisClient()
	if err != nil {
		errorLogger.Error("createRedisClient error", slog.String("msg", err.Error()))
		os.Exit(1)
	}
	infoLogger.Info("Redis Connection", slog.String("Status", "OK"))

	app := &application{
		config:      cfg,
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		models:      data.NewModels(db),
		redisClient: redis,
		done:        make(chan bool),
	}

	err = app.createFakeStockPricesForTesting()
	if err != nil {
		errorLogger.Error("createFakeStockPricesForTesting error", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	err = app.spinUpConsumer()
	if err != nil {
		errorLogger.Error("spinUpConsumer error", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	err = app.serve()
	if err != nil {
		errorLogger.Error("fatal error", slog.String("msg", err.Error()))
		os.Exit(1)
	}
}

func OpenDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duriation, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duriation)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
