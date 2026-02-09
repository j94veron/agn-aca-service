package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"agn-service/internal/cache"
	"agn-service/internal/config"
	"agn-service/internal/db"
	"agn-service/internal/http"
	"agn-service/internal/jobs"
	"agn-service/internal/repository"
	"agn-service/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no se pudo cargar .env (continuo igual):", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	coDB, err := db.NewOracleDB(cfg.OracleCorretaje)
	if err != nil {
		log.Fatal("oracle CO:", err)
	}
	acDB, err := db.NewOracleDB(cfg.OracleAcopio)
	if err != nil {
		log.Fatal("oracle AC:", err)
	}

	rdb := cache.NewRedis(cfg.Redis)
	if err := cache.Ping(ctx, rdb); err != nil {
		log.Fatal("redis:", err)
	}

	coRepo := repository.NewOracleRepo(coDB)
	acRepo := repository.NewOracleRepo(acDB)
	redisRepo := repository.NewRedisRepo(rdb)

	syncJob := jobs.NewSyncJob(coRepo, acRepo, redisRepo, cfg.StatsTTL, cfg.SyncMonths, cfg.ProxWindowMonths)
	svc := service.NewPendingFixService(redisRepo)

	// Cron (segundos)
	cr := cron.New(cron.WithSeconds())
	_, err = cr.AddFunc(cfg.SyncCron, func() {
		if e := syncJob.Run(context.Background()); e != nil {
			log.Println("sync job error:", e)
		} else {
			log.Println("sync job ok")
		}
	})
	if err != nil {
		log.Fatal("cron:", err)
	}
	cr.Start()

	r := http.NewRouter(svc, syncJob)
	log.Println("listening on", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
