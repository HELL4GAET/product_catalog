package dependencies

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"product-catalog/internal/adapters/storage"
	"product-catalog/internal/auth"
	"product-catalog/internal/config"
	"product-catalog/internal/infra/db/pg"
	l "product-catalog/internal/logger"
	"product-catalog/internal/service/file"
	"product-catalog/internal/service/product"
	"product-catalog/internal/service/user"
	h "product-catalog/internal/transport/http"
)

type Deps struct {
	Cfg    *config.Config
	Logger *zap.Logger
	DBPool *pgxpool.Pool

	AuthMiddleware    *auth.Middleware
	LoggingMiddleware *h.LoggingMiddleware

	UserService    *user.Service
	ProductService *product.Service
	FileService    *file.FileService

	UserHandler    *h.UserHandler
	ProductHandler *h.ProductHandler
}

func New(cfg *config.Config) (*Deps, error) {
	// 1. Логгер
	l.Init(cfg.App.Env, cfg.App.LogLevel)
	logger := l.Log

	// 2. Подключение к БД
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	// 3. JWT менеджер и middlewares
	jwtM := auth.NewJWTManager(cfg.JWT.Secret, time.Duration(cfg.JWT.TokenTTLSeconds)*time.Second)
	authM := auth.NewMiddleware(jwtM)
	loggingM := h.NewLoggingMiddleware(logger)

	// 4. Репозитории
	userRepo := pg.NewUserRepo(pool)
	productRepo := pg.NewProductRepo(pool)

	// 5. Сервисы
	hasher := auth.NewHasher()
	userSvc := user.NewUserService(userRepo, hasher, jwtM)
	prodSvc := product.NewProductService(productRepo)

	// 6. Storage (MinIO)
	storageCfg := &storage.MinioConfig{
		Endpoint:       cfg.Storage.Endpoint,
		PublicEndpoint: cfg.Storage.PublicEndpoint,
		AccessKey:      cfg.Storage.AccessKey,
		SecretKey:      cfg.Storage.SecretKey,
		UseSSL:         cfg.Storage.UseSSL,
		Bucket:         cfg.Storage.Bucket,
		Region:         cfg.Storage.Region,
	}
	minioStorage, err := storage.NewMinioStorage(storageCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init MinIO storage: %w", err)
	}
	fileSvc := file.NewFileService(minioStorage)

	// 7. Хендлеры
	userH := h.NewUserHandler(userSvc, logger, authM)
	productH := h.NewProductHandler(prodSvc, fileSvc, logger, authM)

	return &Deps{
		Cfg:               cfg,
		Logger:            logger,
		DBPool:            pool,
		AuthMiddleware:    authM,
		LoggingMiddleware: loggingM,
		UserService:       userSvc,
		ProductService:    prodSvc,
		FileService:       fileSvc,
		UserHandler:       userH,
		ProductHandler:    productH,
	}, nil
}
