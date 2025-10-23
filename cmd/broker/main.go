package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"

	"auth.industrial-linguistics.com/accounting-ops/internal/broker"
)

func main() {
	var (
		envPath = flag.String("env", defaultEnvPath(), "path to broker.env")
		dbPath  = flag.String("db", defaultDBPath(), "path to broker sqlite database")
		addr    = flag.String("addr", ":8080", "listen address when running standalone")
	)
	flag.Parse()

	cfg, err := broker.LoadConfigFromEnvFile(*envPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	store, err := broker.OpenStore(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()

	logger := log.New(os.Stderr, "broker ", log.LstdFlags|log.LUTC)
	server := broker.NewServer(cfg, store, logger)

	if isCGI() {
		logger.Println("running in CGI mode")
		if err := cgi.Serve(server); err != nil {
			logger.Fatalf("cgi serve: %v", err)
		}
		return
	}

	logger.Printf("starting standalone broker on %s", *addr)
	if err := http.ListenAndServe(*addr, server); err != nil {
		logger.Fatalf("listen: %v", err)
	}
}

func isCGI() bool {
	return os.Getenv("GATEWAY_INTERFACE") != ""
}

func defaultEnvPath() string {
	if v := os.Getenv("BROKER_ENV_PATH"); v != "" {
		return v
	}
	return filepath.Join("conf", "broker.env")
}

func defaultDBPath() string {
	if v := os.Getenv("BROKER_DB_PATH"); v != "" {
		return v
	}
	return filepath.Join("data", "broker.sqlite")
}
