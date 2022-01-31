package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ElOtro/stockup-api/internal/data"
	"github.com/jackc/pgx/v4/log/zerologadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Declare a string containing the application version number.
// const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
type config struct {
	port int
	env  string
	seed bool
	db   struct {
		dsn string
	}
	jwt struct {
		secret string
	}
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type application struct {
	config config
	logger *zerolog.Logger
	models data.Models
	seed   data.Seed
}

func main() {
	// Declare an instance of the config struct.
	var cfg config

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}

	// Read the value of the port and env command-line flags into the config struct. We
	// default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")

	// Read the value of the seed and env command-line flags into the config struct. We
	flag.BoolVar(&cfg.seed, "seed", false, "Seed data")

	// Parse the JWT signing secret from the command-line-flag. Notice that we leave the
	// default value as the empty string if no flag is provided.
	flag.StringVar(&cfg.jwt.secret, "jwt-secret", os.Getenv("JWT_SECRET"), "JWT secret")

	flag.Parse()

	// Call the openDB() helper function (see below) to create the connection pool,
	// passing in the config struct. If this returns an error, we log it and exit the
	// application immediately.
	db, err := openDB(cfg)
	if err != nil {
		log.Error().Err(err).Msg("pgx")
	}

	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exits.
	defer db.Close()

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &application{
		config: cfg,
		logger: &logger,
		models: data.NewModels(db),
		seed:   data.Seed{DB: db, Logger: &logger, Models: data.NewModels(db)},
	}

	// generate a `Certificate` struct
	// cert, _ := tls.LoadX509KeyPair("localhost.crt", "localhost.key")

	// Declare a HTTP server with some sensible timeout settings, which listens on the
	// port provided in the config struct and uses the servemux we created above as the
	// handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		// TLSConfig: &tls.Config{
		// 	Certificates: []tls.Certificate{cert},
		// },
	}

	if cfg.seed {
		app.seed.Seed()
	} else {
		// Start the HTTP
		logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
		// err = srv.ListenAndServeTLS("", "")
		err = srv.ListenAndServe()
		log.Fatal().Err(err)
	}

	// // Start the HTTP
	// logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	// err = srv.ListenAndServe()
	// log.Fatal().Err(err)
}

// The openDB() function returns a sql.DB connection pool.
func openDB(cfg config) (*pgxpool.Pool, error) {
	logger := zerologadapter.NewLogger(zerolog.New(os.Stderr).With().Timestamp().Logger())

	poolConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse DATABASE_URL")
		os.Exit(1)
	}

	poolConfig.ConnConfig.Logger = logger

	dbpool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use Ping() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an
	// error.
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	// Return the pgxpool.Pool connection pool.
	return dbpool, nil
}
