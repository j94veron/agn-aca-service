package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv   string
	HTTPAddr string

	OracleCorretaje OracleConfig
	OracleAcopio    OracleConfig
	OracleAGN       OracleConfig

	Redis RedisConfig

	StatsTTL         time.Duration
	SyncCron         string
	SyncCronFix      string
	SyncCronDelivery string
	SyncMonths       int
	ProxWindowMonths int
}

type OracleConfig struct {
	User          string
	Pass          string
	ConnectString string
	LibDir        string
}

type RedisConfig struct {
	Addr string
	Pass string
	DB   int
}

func Load() (Config, error) {
	var cfg Config

	cfg.AppEnv = getenv("APP_ENV", "dev")
	cfg.HTTPAddr = getenv("HTTP_ADDR", ":8080")

	cfg.OracleCorretaje = OracleConfig{
		User:          mustGetenv("ORACLE_CO_USER"),
		Pass:          mustGetenv("ORACLE_CO_PASS"),
		ConnectString: mustGetenv("ORACLE_CO_CONNECT"),
		LibDir:        getenv("ORACLE_CO_LIB_DIR", ""),
	}
	cfg.OracleAcopio = OracleConfig{
		User:          mustGetenv("ORACLE_AC_USER"),
		Pass:          mustGetenv("ORACLE_AC_PASS"),
		ConnectString: mustGetenv("ORACLE_AC_CONNECT"),
		LibDir:        getenv("ORACLE_AC_LIB_DIR", ""),
	}
	cfg.OracleAGN = OracleConfig{
		User:          mustGetenv("ORACLE_AGN_USER"),
		Pass:          mustGetenv("ORACLE_AGN_PASS"),
		ConnectString: mustGetenv("ORACLE_AGN_CONNECT"),
		LibDir:        getenv("ORACLE_AGN_LIB_DIR", ""),
	}

	db, err := strconv.Atoi(getenv("REDIS_DB", "0"))
	if err != nil {
		return cfg, fmt.Errorf("REDIS_DB inválido: %w", err)
	}
	cfg.Redis = RedisConfig{
		Addr: mustGetenv("REDIS_ADDR"),
		Pass: getenv("REDIS_PASS", ""),
		DB:   db,
	}

	ttlHours, err := strconv.Atoi(getenv("STATS_TTL_HOURS", "24"))
	if err != nil {
		return cfg, fmt.Errorf("STATS_TTL_HOURS inválido: %w", err)
	}
	cfg.StatsTTL = time.Duration(ttlHours) * time.Hour

	cfg.SyncCron = getenv("SYNC_CRON_FIX", "0 0 21 * * *")
	cfg.SyncCron = getenv("SYNC_CRON_DELIVERY", "0 0 21 * * *")

	cfg.SyncMonths, err = strconv.Atoi(getenv("SYNC_MONTHS", "12"))
	if err != nil {
		return cfg, fmt.Errorf("SYNC_MONTHS inválido: %w", err)
	}
	cfg.ProxWindowMonths, err = strconv.Atoi(getenv("PROX_WINDOW_MONTHS", "3"))
	if err != nil {
		return cfg, fmt.Errorf("PROX_WINDOW_MONTHS inválido: %w", err)
	}

	return cfg, nil
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("Falta variable de entorno: " + k)
	}
	return v
}
func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
