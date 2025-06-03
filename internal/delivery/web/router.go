package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"ppo/internal/delivery/web/static"

	"ppo/internal/delivery/web/handlers"
	"ppo/internal/services"
)

func NewRouter(
	catSvc services.CategoryService,
	dsSvc services.DatasetService,
	notifSvc services.NotificationService,
	revSvc services.ReviewService,
	subSvc services.SubscriptionService,
	userSvc services.UserService,
	logger *zap.Logger,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Category
	catHandler := handlers.NewCategoryHandler(catSvc, logger)
	catHandler.RegisterRoutes(r)

	// Dataset
	dsHandler := handlers.NewDatasetHandler(dsSvc, logger)
	dsHandler.RegisterRoutes(r)

	// Notification
	notifHandler := handlers.NewNotificationHandler(notifSvc, logger)
	notifHandler.RegisterRoutes(r)

	// Review
	revHandler := handlers.NewReviewHandler(revSvc, logger)
	revHandler.RegisterRoutes(r)

	// Subscription
	subHandler := handlers.NewSubscriptionHandler(subSvc, logger)
	subHandler.RegisterRoutes(r)

	// User
	userHandler := handlers.NewUserHandler(userSvc, logger)
	userHandler.RegisterRoutes(r)

	// Статика (шаблоны + CSS)
	fileServer := http.FileServer(http.FS(static.FS))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return r
}
