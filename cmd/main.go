package main

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"product-catalog/internal/config"
	"product-catalog/internal/di"
)

func main() {
	cfg := config.Load()

	d, err := di.New(cfg)
	if err != nil {
		log.Fatalf("failed to init di: %v", err)
	}
	defer d.DBPool.Close()
	defer d.Logger.Sync()

	r := chi.NewRouter()
	r.Use(d.HTTPMiddleware...)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Mount("/users", d.UserHandler.Routes())
	r.Mount("/products", d.ProductHandler.Routes())

	workDir, _ := os.Getwd()
	docsPath := filepath.Join(workDir, "docs")
	r.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir(docsPath))))

	addr := ":" + cfg.Server.Port
	d.Logger.Info("server started", zap.String("addr", addr))
	d.Logger.Info("OpenAPI spec available at", zap.String("url", "http://localhost"+addr+"/docs/openapi.yaml"))
	if err = http.ListenAndServe(addr, r); err != nil {
		d.Logger.Fatal("failed to start server", zap.Error(err))
	}
}
