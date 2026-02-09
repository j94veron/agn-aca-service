package repository

import (
	"context"
	"encoding/json"
	"time"

	"agn-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	KeyDetail12M  = "agn:stats:v1:pending_fix:detail_12m"
	KeySummary    = "agn:stats:v1:pending_fix:summary"
	KeyMonthly12M = "agn:stats:v1:pending_fix:monthly_12m"
	KeyVencidos   = "agn:stats:v1:pending_fix:vencidos"
)

type RedisRepo struct{ rdb *redis.Client }

func NewRedisRepo(rdb *redis.Client) *RedisRepo { return &RedisRepo{rdb: rdb} }

// generic helpers
func setJSON(ctx context.Context, rdb *redis.Client, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, b, ttl).Err()
}

func getJSON[T any](ctx context.Context, rdb *redis.Client, key string, out *T) error {
	b, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (r *RedisRepo) SaveDetail12M(ctx context.Context, snap domain.PendingFixDetailSnapshot, ttl time.Duration) error {
	return setJSON(ctx, r.rdb, KeyDetail12M, snap, ttl)
}
func (r *RedisRepo) LoadDetail12M(ctx context.Context) (*domain.PendingFixDetailSnapshot, error) {
	var snap domain.PendingFixDetailSnapshot
	if err := getJSON(ctx, r.rdb, KeyDetail12M, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *RedisRepo) SaveSummary(ctx context.Context, snap domain.PendingFixSummarySnapshot, ttl time.Duration) error {
	return setJSON(ctx, r.rdb, KeySummary, snap, ttl)
}
func (r *RedisRepo) LoadSummary(ctx context.Context) (*domain.PendingFixSummarySnapshot, error) {
	var snap domain.PendingFixSummarySnapshot
	if err := getJSON(ctx, r.rdb, KeySummary, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *RedisRepo) SaveMonthly12M(ctx context.Context, snap domain.PendingFixMonthlySnapshot, ttl time.Duration) error {
	return setJSON(ctx, r.rdb, KeyMonthly12M, snap, ttl)
}
func (r *RedisRepo) LoadMonthly12M(ctx context.Context) (*domain.PendingFixMonthlySnapshot, error) {
	var snap domain.PendingFixMonthlySnapshot
	if err := getJSON(ctx, r.rdb, KeyMonthly12M, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func (r *RedisRepo) SaveVencidos(ctx context.Context, snap domain.PendingFixVencidosSnapshot, ttl time.Duration) error {
	return setJSON(ctx, r.rdb, KeyVencidos, snap, ttl)
}
func (r *RedisRepo) LoadVencidos(ctx context.Context) (*domain.PendingFixVencidosSnapshot, error) {
	var snap domain.PendingFixVencidosSnapshot
	if err := getJSON(ctx, r.rdb, KeyVencidos, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}
