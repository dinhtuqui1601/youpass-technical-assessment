package tracking

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Worker consumes view events from Redis Streams and persists them idempotently.
type Worker struct {
	rdb         *redis.Client
	repo        *Repository
	stream      string
	group       string
	consumer    string
	dlqStream   string
	concurrency int
	maxRetries  int
	logger      *slog.Logger
}

// NewWorker builds a view-tracking Worker.
func NewWorker(rdb *redis.Client, repo *Repository, concurrency, maxRetries int, logger *slog.Logger) *Worker {
	return &Worker{
		rdb:         rdb,
		repo:        repo,
		stream:      StreamName,
		group:       GroupName,
		consumer:    "worker-" + uuid.NewString(),
		dlqStream:   DLQStream,
		concurrency: concurrency,
		maxRetries:  maxRetries,
		logger:      logger,
	}
}

// Run starts the consumer group, the pending-entry reclaimer and a bounded
// worker pool, then blocks until the context is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	if err := w.ensureGroup(ctx); err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.reclaimLoop(ctx)
	}()

	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.consume(ctx)
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

func (w *Worker) ensureGroup(ctx context.Context) error {
	err := w.rdb.XGroupCreateMkStream(ctx, w.stream, w.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (w *Worker) consume(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		streams, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    w.group,
			Consumer: w.consumer,
			Streams:  []string{w.stream, ">"},
			Count:    10,
			Block:    5 * time.Second,
			NoAck:    false,
		}).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if w.logger != nil {
				w.logger.Error("xreadgroup failed", "error", err)
			}
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				w.process(ctx, stream.Stream, msg)
			}
		}
	}
}

// reclaimLoop periodically claims pending entries (e.g. left behind by a crash)
// and re-processes them.
func (w *Worker) reclaimLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.reclaimPending(ctx)
		}
	}
}

func (w *Worker) reclaimPending(ctx context.Context) {
	start := "0"
	for {
		entries, next, err := w.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   w.stream,
			Group:    w.group,
			Consumer: w.consumer,
			MinIdle:  60 * time.Second,
			Start:    start,
			Count:    10,
		}).Result()
		if err != nil {
			if w.logger != nil {
				w.logger.Warn("xautoclaim failed", "error", err)
			}
			return
		}
		for _, msg := range entries {
			w.process(ctx, w.stream, msg)
		}
		if len(entries) == 0 || next == "0-0" || next == "" {
			return
		}
		start = next
	}
}

// process handles a single stream message with bounded retries and a dead
// letter queue. It acks only after a successful (or idempotent) insert.
func (w *Worker) process(ctx context.Context, stream string, msg redis.XMessage) {
	shareLinkID, err := primitive.ObjectIDFromHex(toString(msg.Values["shareLinkId"]))
	if err != nil {
		w.ack(ctx, stream, msg.ID) // malformed message, drop it
		return
	}

	viewID := toString(msg.Values["viewId"])
	viewedAt, err := time.Parse(time.RFC3339Nano, toString(msg.Values["viewedAt"]))
	if err != nil {
		viewedAt = time.Now().UTC()
	}
	retries := toInt(msg.Values["retries"])

	rec := &ViewRecord{
		ShareLinkID: shareLinkID,
		ViewID:      viewID,
		ViewedAt:    viewedAt,
		IP:          toString(msg.Values["ip"]),
		UserAgent:   toString(msg.Values["userAgent"]),
		CreatedAt:   time.Now().UTC(),
	}

	err = w.repo.Record(ctx, rec)
	if err == nil || errors.Is(err, ErrDuplicateView) {
		w.ack(ctx, stream, msg.ID)
		return
	}

	if retries >= w.maxRetries {
		// Move to the dead-letter stream, then ack the original.
		if err := w.rdb.XAdd(ctx, &redis.XAddArgs{Stream: w.dlqStream, Values: msg.Values}).Err(); err != nil && w.logger != nil {
			w.logger.Error("dlq add failed", "error", err)
		}
		w.ack(ctx, stream, msg.ID)
		return
	}

	// Re-enqueue with retries+1 BEFORE acking so a crash cannot lose the event;
	// the unique (shareLinkId, viewId) index keeps the count correct.
	next := cloneValues(msg.Values)
	next["retries"] = strconv.Itoa(retries + 1)
	if err := w.rdb.XAdd(ctx, &redis.XAddArgs{Stream: w.stream, Values: next}).Err(); err != nil {
		if w.logger != nil {
			w.logger.Error("retry add failed", "error", err)
		}
		return // leave unacked; the reclaimer will retry it later
	}
	w.ack(ctx, stream, msg.ID)
}

func (w *Worker) ack(ctx context.Context, stream, id string) {
	if err := w.rdb.XAck(ctx, stream, w.group, id).Err(); err != nil && w.logger != nil {
		w.logger.Warn("ack failed", "stream", stream, "id", id, "error", err)
	}
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}

func toInt(v any) int {
	s := toString(v)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func cloneValues(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

