package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/uvalib/easystore/uvaeasystore"
)

// ServiceConfig defines the service configuration parameters
type ServiceConfig struct {
	Port  int
	esCfg uvaeasystore.DatastoreS3Config
	//esCfg uvaeasystore.DatastorePostgresConfig
}

func envWithDefault(env string, defaultValue string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("INFO: environment variable not set: [%s] using default value [%s]", env, defaultValue)
		return defaultValue
	}

	return val
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

	cfg.Port = envToInt("ES_SERVICE_PORT")

	cfg.esCfg.Bucket = ensureSetAndNonEmpty("ES_BUCKET")
	cfg.esCfg.SignerAccessKey = envWithDefault("SIGNER_ACCESS_KEY", "")
	cfg.esCfg.SignerSecretKey = envWithDefault("SIGNER_SECRET_KEY", "")
	cfg.esCfg.SignerExpireMinutes = envToInt("ES_SIGN_EXPIRE_MINUTES")

	cfg.esCfg.DbHost = ensureSetAndNonEmpty("ES_DBHOST")
	cfg.esCfg.DbPort = envToInt("ES_DBPORT")
	cfg.esCfg.DbName = ensureSetAndNonEmpty("ES_DBNAME")
	cfg.esCfg.DbUser = ensureSetAndNonEmpty("ES_DBUSER")
	cfg.esCfg.DbPassword = ensureSetAndNonEmpty("ES_DBPASS")
	cfg.esCfg.DbTimeout = envToInt("ES_DBTIMEOUT")
	cfg.esCfg.BusName = ensureSetAndNonEmpty("ES_BUS_NAME")
	cfg.esCfg.SourceName = ensureSetAndNonEmpty("ES_SOURCE_NAME")

	log.Printf("[CONFIG] Port                = [%d]", cfg.Port)

	log.Printf("[CONFIG] Bucket              = [%s]", cfg.esCfg.Bucket)
	log.Printf("[CONFIG] SignerAccessKey     = [%s]", strings.Repeat("*", len(cfg.esCfg.SignerAccessKey)))
	log.Printf("[CONFIG] SignerSecretKey     = [%s]", strings.Repeat("*", len(cfg.esCfg.SignerAccessKey)))
	log.Printf("[CONFIG] SignerExpireMinutes = [%d]", cfg.esCfg.SignerExpireMinutes)

	log.Printf("[CONFIG] DbHost              = [%s]", cfg.esCfg.DbHost)
	log.Printf("[CONFIG] DbPort              = [%d]", cfg.esCfg.DbPort)
	log.Printf("[CONFIG] DbName              = [%s]", cfg.esCfg.DbName)
	log.Printf("[CONFIG] DbUser              = [%s]", cfg.esCfg.DbUser)
	log.Printf("[CONFIG] DbPassword          = [%s]", strings.Repeat("*", len(cfg.esCfg.DbPassword)))
	log.Printf("[CONFIG] DbTimeout           = [%d]", cfg.esCfg.DbTimeout)
	log.Printf("[CONFIG] BusName             = [%s]", cfg.esCfg.BusName)
	log.Printf("[CONFIG] SourceName          = [%s]", cfg.esCfg.SourceName)

	cfg.esCfg.Log = log.Default()

	return &cfg
}
