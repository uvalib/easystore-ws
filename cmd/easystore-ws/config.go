package main

import (
	"log"
	"os"
	"strconv"
)

// ServiceConfig defines all of the archives transfer service configuration paramaters
type ServiceConfig struct {
	SharedSecret string
	ExpireDays   int
	Port         int
}

func ensureSet(env string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Fatalf("environment variable not set: [%s]", env)
	}

	return val
}

func ensureSetAndNonEmpty(env string) string {
	val := ensureSet(env)

	if val == "" {
		log.Fatalf("environment variable not set: [%s]", env)
	}

	return val
}

func envToInt(env string) int {

	number := ensureSetAndNonEmpty(env)
	n, err := strconv.Atoi(number)
	if err != nil {
		log.Fatalf("cannot convert to integer: [%s]", env)
	}
	return n
}

// LoadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func LoadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	cfg.SharedSecret = ensureSetAndNonEmpty("MINT_TOKEN_SHARED_SECRET")
	cfg.ExpireDays = envToInt("MINT_TOKEN_EXPIRE_DAYS")
	cfg.Port = envToInt("MINT_TOKEN_SERVICE_PORT")

	log.Printf("[CONFIG] SharedSecret  = [REDACTED]")
	log.Printf("[CONFIG] ExpireDays    = [%d]", cfg.ExpireDays)
	log.Printf("[CONFIG] Port          = [%d]", cfg.Port)

	return &cfg
}
