package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ayushpandey15/lambda-go/internal/config"
	"github.com/ayushpandey15/lambda-go/internal/handler"
	"github.com/ayushpandey15/lambda-go/internal/router"
	"github.com/ayushpandey15/lambda-go/internal/server"
	"github.com/gin-gonic/gin"

	_ "github.com/ayushpandey15/lambda-go/internal/router/health" // Register routes (init → router.Register)
	_ "github.com/ayushpandey15/lambda-go/internal/router/pdf"
	"github.com/ayushpandey15/lambda-go/internal/router/middleware/auth"
)

func main() {
	if _, err := config.Load(context.Background()); err != nil {
		log.Fatal(err)
	}

	router.ApplicationStartTime = time.Now()
	engine := server.NewEngine()
	setupMiddleware(engine)

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		gin.SetMode(gin.ReleaseMode)
		lambda.Start(handler.APIGateway(engine))
		return
	}

	runLocal(engine)
}

func runLocal(engine *gin.Engine) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := config.ListenAddr()
	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("gin listening on %s (env=%s)", addr, config.Env())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	go heartbeat(ctx)

	<-ctx.Done()
	log.Println("shutting down…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func heartbeat(ctx context.Context) {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			log.Printf("heartbeat env=%s uptime=%s", config.Env(), time.Since(router.ApplicationStartTime).Round(time.Second))
		}
	}
}

func setupMiddleware(engine *gin.Engine) {
	engine.Use(auth.AuthCheck)
}
