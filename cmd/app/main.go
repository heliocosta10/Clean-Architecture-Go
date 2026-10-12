package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"orders-clean-architecture/internal/adapter/graphqlapi"
	"orders-clean-architecture/internal/adapter/postgres"
	"orders-clean-architecture/internal/adapter/rest"
	"orders-clean-architecture/internal/adapter/rpc"
	"orders-clean-architecture/internal/usecase"
)

func main() {
	databaseURL := env("DATABASE_URL", "postgres://orders:orders@localhost:5432/orders?sslmode=disable")
	pool, err := connectWithRetry(databaseURL, 30, time.Second)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	if err := applyMigrations(context.Background(), pool); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	repo := postgres.NewOrderRepository(pool)
	listOrders := usecase.NewListOrders(repo)
	createOrder := usecase.NewCreateOrder(repo)

	restHandler := rest.NewHandler(listOrders, createOrder)
	mux := http.NewServeMux()
	mux.Handle("/order", restHandler)
	mux.Handle("/order/", restHandler)

	gqlSchema, err := graphqlapi.NewSchema(listOrders, createOrder)
	if err != nil {
		log.Fatalf("create GraphQL schema: %v", err)
	}
	mux.Handle("/graphql", graphqlapi.Handler(gqlSchema))

	httpServer := &http.Server{Addr: ":" + env("WEB_PORT", "8080"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("REST listening on :%s; GraphQL on /graphql", env("WEB_PORT", "8080"))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	rpc.Register(grpcServer, listOrders)
	reflection.Register(grpcServer)
	listener, err := net.Listen("tcp", ":"+env("GRPC_PORT", "50051"))
	if err != nil {
		log.Fatalf("listen gRPC: %v", err)
	}
	log.Printf("gRPC listening on :%s", env("GRPC_PORT", "50051"))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("gRPC server: %v", err)
	}
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationPaths := []string{"migrations", "/app/migrations"}
	var dir string
	for _, candidate := range migrationPaths {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			dir = candidate
			break
		}
	}
	if dir == "" {
		return errors.New("migrations directory not found")
	}

	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		var alreadyApplied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, file.Name()).Scan(&alreadyApplied); err != nil {
			return err
		}
		if alreadyApplied {
			log.Printf("migration %s already applied; skipping", file.Name())
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, string(content)); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, file.Name())
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return errors.New("apply migration " + file.Name() + ": " + err.Error())
		}
		if err := tx.Commit(ctx); err != nil {
			return errors.New("commit migration " + file.Name() + ": " + err.Error())
		}
		log.Printf("applied migration %s", file.Name())
	}
	return nil
}

func connectWithRetry(url string, attempts int, delay time.Duration) (*pgxpool.Pool, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pool, err := pgxpool.New(ctx, url)
		if err == nil {
			err = pool.Ping(ctx)
		}
		cancel()
		if err == nil {
			return pool, nil
		}
		lastErr = err
		if pool != nil {
			pool.Close()
		}
		time.Sleep(delay)
	}
	return nil, lastErr
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
