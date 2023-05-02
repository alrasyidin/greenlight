package main

import (
	"context"
	"database/sql"
	"expvar"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
	// Import the pq driver so that it can register itself with the database/sql

	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/DataDavD/snippetbox/greenlight/internal/data"
	"github.com/DataDavD/snippetbox/greenlight/internal/jsonlog"
	"github.com/DataDavD/snippetbox/greenlight/internal/mailer"
)

// Declare a string containing the application version number
const version = "1.0.0"

// Define an application struct to hold dependencies for our HTTP handlers, helpers, and
// middleware.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	// Initialize a new jsonlog.Logger which writes any messages *at or above* the INFO
	// severity level to the standard out stream.
	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	// Load env variable
	err := godotenv.Load()

	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Declare an instance of the config struct.
	var cfg config
	// load all config
	cfg.load()

	// if display version true exit from app
	if cfg.displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	// Call the openDB() helper function (see below) to create teh connection pool,
	// passing in the config struct. If this returns an error,
	// we log it and exit the application immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main()
	// function exits.
	defer func() {
		if err := db.Close(); err != nil {
			logger.PrintFatal(err, nil)
		}
	}()

	logger.PrintInfo("database connection pool established", nil)

	// Publish a new "version" varaible in the expar var handler containing our application
	// version number.
	expvar.NewString("version").Set(version)

	// Publish the number of activate goroutines.
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Declare an instance of the application struct, containing the config struct and the infoLog.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// Call app.server() to start the server.
	if err := app.serve(); err != nil {
		logger.PrintFatal(err, nil)
	}
}

// openDB returns a sql.DB connection pool to postgres database
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	// Note that passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connection in the pool. Again,
	// passing a value less than or equal to 0 will mean there is no limit
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout duration string to a
	// time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database,
	// passing in the context we created above as a parameter.
	// If connection couldn't be established successfully within the 5-second deadline,
	// then this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}
