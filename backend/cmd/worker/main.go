package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/youpass/backend/configs"
	"github.com/youpass/backend/internal/tracking"
	"github.com/youpass/backend/pkg/database"
	"github.com/youpass/backend/pkg/logger"
	"github.com/youpass/backend/pkg/redis"
)

const workerMaxRetries = 5

func main() {
	cfg := configs.Load()
	log := logger.New()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()

	mongoClient, err := database.Connect(startupCtx, cfg.MongoURI)
	if err != nil {
		log.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database(cfg.MongoDatabase)
	if err := database.EnsureIndexes(startupCtx, db); err != nil {
		log.Error("failed to ensure indexes", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer rdb.Close()

	repo := tracking.NewRepository(db)
	worker := tracking.NewWorker(rdb, repo, cfg.ViewWorkerConcurrency, workerMaxRetries, log)

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("worker: starting", "concurrency", cfg.ViewWorkerConcurrency)
	if err := worker.Run(sigCtx); err != nil {
		log.Error("worker: exited with error", "error", err)
		os.Exit(1)
	}
	log.Info("worker: stopped")
}

