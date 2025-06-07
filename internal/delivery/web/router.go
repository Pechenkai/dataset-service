package web

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	webmid "ppo/internal/delivery/web/middleware"
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
	r.Use(webmid.ChiZapLogger(logger))

	r.Use(webmid.AuthMiddleware)

	// Category
	catHandler := handlers.NewCategoryHandler(catSvc, dsSvc, logger)
	catHandler.RegisterRoutes(r)

	// Dataset
	dsHandler := handlers.NewDatasetHandler(dsSvc, revSvc, userSvc, catSvc, subSvc, notifSvc, logger)
	dsHandler.RegisterRoutes(r)

	// Notification
	notifHandler := handlers.NewNotificationHandler(notifSvc, dsSvc, logger)
	notifHandler.RegisterRoutes(r)

	// Review
	revHandler := handlers.NewReviewHandler(revSvc, logger)
	revHandler.RegisterRoutes(r)

	// Subscription
	subHandler := handlers.NewSubscriptionHandler(subSvc, logger)
	subHandler.RegisterRoutes(r)

	// User
	userHandler := handlers.NewUserHandler(userSvc, notifSvc, logger)
	userHandler.RegisterRoutes(r)

	// Статика (шаблоны + CSS)
	fileServer := http.FileServer(http.FS(static.FS))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return r
}
