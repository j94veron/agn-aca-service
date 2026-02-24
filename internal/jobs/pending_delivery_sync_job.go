package jobs

import (
	"context"
	"sync"
	"time"

	"agn-service/internal/domain"
	"agn-service/internal/repository"
)

type PendingDeliverySyncJob struct {
	coRepo *repository.OraclePendingEntregaRepo
	acRepo *repository.OraclePendingEntregaRepo
	redis  *repository.RedisRepo
	ttl    time.Duration
}

func NewPendingDeliverySyncJob(
	coRepo *repository.OraclePendingEntregaRepo,
	acRepo *repository.OraclePendingEntregaRepo,
	redis *repository.RedisRepo,
	ttl time.Duration,
) *PendingDeliverySyncJob {
	return &PendingDeliverySyncJob{
		coRepo: coRepo,
		acRepo: acRepo,
		redis:  redis,
		ttl:    ttl,
	}
}

func (j *PendingDeliverySyncJob) Run(ctx context.Context) error {
	now := time.Now()

	type res struct {
		rows []domain.PendingDeliveryRow
		err  error
	}

	var wg sync.WaitGroup
	chCO := make(chan res, 1)
	chAC := make(chan res, 1)

	wg.Add(2)

	go func() {
		defer wg.Done()
		rows, err := j.coRepo.FetchAll(ctx, "CORRETAJE", "CO")
		chCO <- res{rows: rows, err: err}
	}()

	go func() {
		defer wg.Done()
		rows, err := j.acRepo.FetchAll(ctx, "ACOPIO", "AC")
		chAC <- res{rows: rows, err: err}
	}()

	wg.Wait()

	r1 := <-chCO
	r2 := <-chAC

	if r1.err != nil {
		return r1.err
	}
	if r2.err != nil {
		return r2.err
	}

	all := make([]domain.PendingDeliveryRow, 0, len(r1.rows)+len(r2.rows))
	all = append(all, r1.rows...)
	all = append(all, r2.rows...)

	snap := domain.PendingDeliverySnapshot{
		GeneratedAt: now,
		Rows:        all,
	}

	return j.redis.SavePendingDelivery(ctx, snap, j.ttl)
}
