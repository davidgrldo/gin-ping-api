package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Handle SIGINT / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ─── Gin router ─────────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": GenerateTimestamp(),
		})
	})

	// ─── HTTP server with graceful shutdown ─────────────────────────────────────
	addr := ":" + getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Wait for signal
	<-ctx.Done()
	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

// getEnv returns fallback if key is unset / empty
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GenerateTimestamp() string {
	// Load Jakarta timezone (GMT+7)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// fallback to GMT+7 fixed zone if system doesn't have timezone data
		loc = time.FixedZone("GMT+7", 7*60*60)
	}

	return time.Now().In(loc).Format("2006-01-02T15:04:05.000+07")
}
