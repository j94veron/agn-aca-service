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
	"agn-service/internal/oracle"
	"agn-service/internal/repository"
	"agn-service/internal/service"
)

func main() {
	// ---- ENV ----
	if err := godotenv.Load(); err != nil {
		log.Println("no se pudo cargar .env (continuo igual):", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config:", err)
	}

	ctx := context.Background()
	log.Println("service starting...")

	// ---- ORACLE ----
	coDB, err := db.NewOracleDB(cfg.OracleCorretaje)
	if err != nil {
		log.Fatal("oracle CO:", err)
	}

	acDB, err := db.NewOracleDB(cfg.OracleAcopio)
	if err != nil {
		log.Fatal("oracle AC:", err)
	}

	agnDB, err := db.NewOracleDB(cfg.OracleAGN)
	if err != nil {
		log.Fatal("oracle AGN:", err)
	}

	log.Println("oracle connections OK")

	// ---- REDIS ----
	rdb := cache.NewRedis(cfg.Redis)
	if err := cache.Ping(ctx, rdb); err != nil {
		log.Fatal("redis:", err)
	}

	log.Println("redis connection OK")

	// ---- REPOSITORIES ----
	coRepo := repository.NewOracleRepo(coDB)
	acRepo := repository.NewOracleRepo(acDB)

	agnClient := oracle.NewFromSqlx(agnDB)

	contratoRepo := repository.NewContratoRepo(agnClient)
	oliquiRepo := repository.NewOliquiRepo(agnClient)
	ctamreRepo := repository.NewCtamreRepo(agnClient)
	movauditRepo := repository.NewMovAuditRepo(agnClient)

	redisRepo := repository.NewRedisRepo(rdb)

	// ---- SERVICES AGN ----
	val := service.NewValidators(
		contratoRepo,
		oliquiRepo,
		ctamreRepo,
	)

	oliquiSvc := service.NewOliquiService(
		agnClient,
		val,
		contratoRepo,
		oliquiRepo,
		movauditRepo,
		oliquiRepo,
	)

	ctamreSvc := service.NewCtamreService(
		agnClient,
		val,
		contratoRepo,
		ctamreRepo,
		movauditRepo,
	)

	// ---- PENDING FIX ----
	syncJob := jobs.NewSyncJob(
		coRepo,
		acRepo,
		redisRepo,
		cfg.StatsTTL,
		cfg.SyncMonths,
		cfg.ProxWindowMonths,
	)

	pendingFixSvc := service.NewPendingFixService(redisRepo)

	// ---- PENDING DELIVERY ----
	pendingDeliveryRepoCO := repository.NewOraclePendingEntregaRepo(coDB)
	pendingDeliveryRepoAC := repository.NewOraclePendingEntregaRepo(acDB)

	pendingDeliverySyncJob := jobs.NewPendingDeliverySyncJob(
		pendingDeliveryRepoCO,
		pendingDeliveryRepoAC,
		redisRepo,
		cfg.StatsTTL,
	)

	pendingDeliverySvc := service.NewPendingDeliveryService(redisRepo)

	// ---- CRON JOBS ----
	cr := cron.New(cron.WithSeconds())

	if cfg.SyncCronFix != "" {
		_, err = cr.AddFunc(cfg.SyncCronFix, func() {
			if e := syncJob.Run(ctx); e != nil {
				log.Println("sync job error:", e)
			} else {
				log.Println("sync job ok")
			}
		})
		if err != nil {
			log.Fatal("cron fix:", err)
		}
	}

	if cfg.SyncCronDelivery != "" {
		_, err = cr.AddFunc(cfg.SyncCronDelivery, func() {
			if e := pendingDeliverySyncJob.Run(ctx); e != nil {
				log.Println("pending delivery sync error:", e)
			} else {
				log.Println("pending delivery sync ok")
			}
		})
		if err != nil {
			log.Fatal("cron delivery:", err)
		}
	}

	cr.Start()
	log.Println("cron jobs started")

	// ---- HTTP SERVER ----
	r := http.NewRouter(
		pendingFixSvc,
		pendingDeliverySvc,
		syncJob,
		pendingDeliverySyncJob,
		oliquiSvc,
		ctamreSvc,
	)

	log.Println("HTTP listening on", cfg.HTTPAddr)

	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal("http server:", err)
	}
}
