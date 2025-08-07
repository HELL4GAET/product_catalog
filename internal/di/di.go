package di

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"net/http"
	"product-catalog/internal/adapters/storage"
	"product-catalog/internal/infra/db/pg"
	"product-catalog/internal/service/file"
	"time"

	"product-catalog/internal/auth"
	"product-catalog/internal/config"
	l "product-catalog/internal/logger"
	"product-catalog/internal/service/product"
	"product-catalog/internal/service/user"
	h "product-catalog/internal/transport/http"
)

type Deps struct {
	Cfg    *config.Config
	Logger *zap.Logger
	DBPool *pgxpool.Pool

	AuthMiddleware *auth.Middleware
	HTTPMiddleware []func(http.Handler) http.Handler

	UserService    *user.Service
	ProductService *product.Service
	FileService    *file.FileService

	UserHandler    *h.UserHandler
	ProductHandler *h.ProductHandler
}

func New(cfg *config.Config) (*Deps, error) {
	l.Init(cfg.App.Env, cfg.App.LogLevel)
	logger := l.Log

	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		logger.Fatal("failed to connect to db", zap.Error(err))
		return nil, err
	}

	jwtM := auth.NewJWTManager(cfg.JWT.Secret, time.Duration(cfg.JWT.TokenTTLSeconds)*time.Second)
	authM := auth.NewMiddleware(jwtM)
	loggingM := h.NewLoggingMiddleware(logger)

	userRepo := pg.NewUserRepo(pool)
	productRepo := pg.NewProductRepo(pool)

	hasher := auth.NewHasher()
	userSvc := user.NewUserService(userRepo, hasher, jwtM)
	prodSvc := product.NewProductService(productRepo)

	storageCfg := storage.NewConfig(cfg)
	minioStorage, err := storage.NewMinioStorage(storageCfg)
	if err != nil {
		logger.Fatal("failed to init minio storage", zap.Error(err))
	}
	fileSvc := file.NewFileService(minioStorage)

	userH := h.NewUserHandler(userSvc, logger)
	productH := h.NewProductHandler(prodSvc, fileSvc, logger)

	httpMiddles := []func(handler http.Handler) http.Handler{
		loggingM.LoggingMiddleware,
		authM.AuthMiddleware,
	}

	return &Deps{
		Cfg:            cfg,
		Logger:         logger,
		DBPool:         pool,
		AuthMiddleware: authM,
		HTTPMiddleware: httpMiddles,
		UserService:    userSvc,
		ProductService: prodSvc,
		FileService:    fileSvc,
		UserHandler:    userH,
		ProductHandler: productH,
	}, nil

}
