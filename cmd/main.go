package main

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
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

	r.Mount("/users", d.UserHandler.Routes())
	r.Mount("products", d.ProductHandler.Routes())

	addr := ":" + cfg.Server.Port
	d.Logger.Info("server started", zap.String("addr", addr))
	if err = http.ListenAndServe(addr, r); err != nil {
		d.Logger.Fatal("failed to start server", zap.Error(err))
	}
}
