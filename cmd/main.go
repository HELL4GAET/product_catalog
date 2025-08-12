package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"product-catalog/internal/config"
	"product-catalog/internal/dependencies"
)

func main() {
	cfg := config.Load()

	d, err := dependencies.New(cfg)
	if err != nil {
		log.Fatalf("failed to init dependencies: %v", err)
	}
	defer d.DBPool.Close()
	defer d.Logger.Sync()

	r := chi.NewRouter()
	r.Use(d.LoggingMiddleware.LoggingMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Mount("/users", d.UserHandler.Routes())
	r.Mount("/products", d.ProductHandler.Routes())

	r.Route("/docs", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:8080"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
		workDir, _ := os.Getwd()
		docsPath := filepath.Join(workDir, "docs")
		r.Handle("/*", http.StripPrefix("/docs/", http.FileServer(http.Dir(docsPath))))
	})

	addr := ":" + cfg.Server.Port
	d.Logger.Info("server started", zap.String("addr", addr))
	d.Logger.Info("OpenAPI spec available at", zap.String("url", "http://localhost"+addr+"/docs/openapi.yaml"))
	if err = http.ListenAndServe(addr, r); err != nil {
		d.Logger.Fatal("failed to start server", zap.Error(err))
	}
}
