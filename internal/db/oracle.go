package db

import (
	"fmt"
	"os"

	_ "github.com/godror/godror"
	"github.com/jmoiron/sqlx"

	"agn-service/internal/config"
)

func NewOracleDB(cfg config.OracleConfig) (*sqlx.DB, error) {
	if cfg.LibDir != "" {
		_ = os.Setenv("LD_LIBRARY_PATH", cfg.LibDir)
	}
	dsn := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`,
		cfg.User, cfg.Pass, cfg.ConnectString)

	db, err := sqlx.Open("godror", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
