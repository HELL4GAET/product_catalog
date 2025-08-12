package http

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"product-catalog/internal/auth"
	"product-catalog/internal/domain"
	"strconv"
	"time"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domain.Product) (int, error)
	GetProductByID(ctx context.Context, id int) (*domain.Product, error)
	GetAllProducts(ctx context.Context) ([]domain.Product, error)
	UpdateProductByID(ctx context.Context, id int, product *domain.Product) error
	DeleteProductByID(ctx context.Context, id int) error
}

type FileService interface {
	Upload(ctx context.Context, fh *multipart.FileHeader) (string, error)
}

type ProductHandler struct {
	productSvc     ProductService
	fileSvc        FileService
	logger         *zap.Logger
	authMiddleware func(http.Handler) http.Handler
}

func NewProductHandler(productCvc ProductService, fileSvc FileService, logger *zap.Logger, authMiddleware *auth.Middleware) *ProductHandler {
	return &ProductHandler{productSvc: productCvc, fileSvc: fileSvc, logger: logger, authMiddleware: authMiddleware.AuthMiddleware}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)
		r.Post("/", h.CreateProduct)
		r.Put("/{id}", h.UpdateProductByID)
		r.Delete("/{id}", h.DeleteProductByID)
	})

	r.Get("/", h.GetAllProducts)
	r.Get("/{id}", h.GetProductByID)
	return r
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.logger.Warn("failed to parse multipart form", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		h.logger.Warn("invalid request", zap.String("title", title))
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	priceStr := r.FormValue("price")
	if priceStr == "" {
		h.logger.Warn("invalid request", zap.String("price", priceStr))
		http.Error(w, "price is required", http.StatusBadRequest)
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if priceStr == "" {
		h.logger.Warn("invalid request", zap.String("price", priceStr))
		http.Error(w, "price is required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		h.logger.Warn("invalid request", zap.String("description", description))
		http.Error(w, "description is required", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["image"]
	if len(files) == 0 {
		h.logger.Warn("request without image")
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	fileHeader := files[0]

	key, err := h.fileSvc.Upload(r.Context(), fileHeader)
	if err != nil {
		h.logger.Error("failed to upload image", zap.Error(err))
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}

	//url, err := h.fileSvc.GetPresignedURL(r.Context(), key)
	//if err != nil {
	//	h.logger.Error("failed to get presigned url", zap.Error(err))
	//	http.Error(w, "upload error", http.StatusInternalServerError)
	//	return
	//}

	prod := &domain.Product{
		Title:       title,
		Price:       int(price),
		Description: description,
		ImageURL:    key,
		CreatedAt:   time.Now(),
	}

	id, err := h.productSvc.CreateProduct(r.Context(), prod)
	if err != nil {
		h.logger.Error("failed to create product", zap.Error(err))
		http.Error(w, "creation error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]int{"id": id})
	if err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productSvc.GetAllProducts(r.Context())
	if err != nil {
		h.logger.Error("failed to get products", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(products)
	if err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.logger.Warn("invalid request with empty id", zap.String("id", idStr))
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("invalid product id", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	product, err := h.productSvc.GetProductByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get product", zap.Error(err))
		http.Error(w, "get product error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(product)
	if err != nil {
		h.logger.Warn("failed to encode response", zap.Error(err))
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}

func (h *ProductHandler) UpdateProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.logger.Warn("invalid request with empty id", zap.String("id", idStr))
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("invalid product id", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "invalid product id", http.StatusBadRequest)
	}

	if err = r.ParseMultipartForm(10 << 20); err != nil {
		h.logger.Warn("failed to parse multipart form", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		h.logger.Warn("invalid request", zap.String("title", title))
		http.Error(w, "title is required", http.StatusBadRequest)
	}

	priceStr := r.FormValue("price")
	if priceStr == "" {
		h.logger.Warn("invalid request", zap.String("price", priceStr))
		http.Error(w, "price is required", http.StatusBadRequest)
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if priceStr == "" {
		h.logger.Warn("invalid request", zap.String("price", priceStr))
		http.Error(w, "price is required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		h.logger.Warn("invalid request", zap.String("description", description))
		http.Error(w, "description is required", http.StatusBadRequest)
	}

	availableStr := r.FormValue("available")
	available, err := strconv.ParseBool(availableStr)
	if err != nil {
		available = false
	}

	files := r.MultipartForm.File["image"]
	if len(files) != 0 {
		fileHeader := files[0]
		imageURL, err := h.fileSvc.Upload(r.Context(), fileHeader)
		if err != nil {
			h.logger.Error("failed to upload image", zap.Error(err))
			http.Error(w, "upload error", http.StatusInternalServerError)
			return
		}
		product := &domain.Product{
			Title:       title,
			Price:       int(price),
			Description: description,
			Available:   available,
			ImageURL:    imageURL,
		}
		err = h.productSvc.UpdateProductByID(r.Context(), id, product)
		if err != nil {
			h.logger.Error("failed to update product", zap.Error(err))
			http.Error(w, "update error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	product := &domain.Product{
		Title:       title,
		Price:       int(price),
		Description: description,
		Available:   available,
	}
	err = h.productSvc.UpdateProductByID(r.Context(), id, product)
	if err != nil {
		h.logger.Error("failed to update product", zap.Error(err))
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func (h *ProductHandler) DeleteProductByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.logger.Warn("invalid request with empty id", zap.String("id", idStr))
		http.Error(w, "id is required", http.StatusBadRequest)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("invalid product id", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "invalid product id", http.StatusBadRequest)
	}
	err = h.productSvc.DeleteProductByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete product", zap.Error(err))
		http.Error(w, "delete error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
