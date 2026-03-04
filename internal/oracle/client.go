package oracle

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Client struct {
	DB *sql.DB
}

func NewFromSqlx(db *sqlx.DB) *Client {
	return &Client{
		DB: db.DB,
	}
}

func (c *Client) Close() error {
	return c.DB.Close()
}

func (c *Client) CtxWithTimeout(sec int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(sec)*time.Second)
}
