package main

import (
	"flag"
	"os"
	"strings"
)

// Define a config struct.
type config struct {
	port           int
	env            string
	displayVersion bool
	// db struct field holds the configuration settings for our database connection pool.
	// For now this only holds the DSN, which we read in from a command-line flag.
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	// Add a new limiter struct containing fields for the request-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

func (cfg *config) load() {
	// Read the value of the port and env command-line flags into the config struct.
	// We default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", getenvInt("PORT"), "API server port")
	flag.StringVar(&cfg.env, "env", getenvStr("ENV"), "Environment (development|staging|production")

	// Read the DSN Value from the db-dsn command-line flag into the config struct.
	// We default to using our development DSN if no flag is provided.
	// pw := os.Getenv("DB_PW")
	flag.StringVar(&cfg.db.dsn, "db-dsn", getenvStr("DB_DSN"), "PostgreSQL DSN")
	// Read the connection pool settings from command-line flags into the config struct.
	// Notice the default values that we're using?
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", getenvInt("DB_MAX_OPEN_CONNS"),
		"PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", getenvInt("DB_MAX_IDLE_CONNS"),
		"PostgreSQL max open idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", getenvStr("DB_MAX_IDLE_TIME"),
		"PostgreSQL max connection idle time")

	// Read the limiter settings from the command-line flags into the config struct.
	// We use true as the default for 'enabled' setting.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as teh default values.
	mtUser := os.Getenv("MAILTRAP_USER")
	mtPw := os.Getenv("MAILTRAP_PW")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", mtUser, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", mtPw, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "DoNotReply <3fc3f54366-09689f+1@inbox.mailtrap.io>", "SMTP sender")

	// Use flag.Func function to process the -cors-trusted-origins command line flag. In this we
	// use the strings.Field function to split the flag value into slice based on whitespace
	// characters and assign it to our config struct. Importantly, if the -cors-trusted-origins
	// flag is not present, contains the empty string, or contains only whitespace, then
	// strings.Fields will return an empty []string slice.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.BoolVar(&cfg.displayVersion, "version", false, "Display version and exit")

	flag.Parse()
}
