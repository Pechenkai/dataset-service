package delivhttp

import (
	"go.uber.org/zap"
	"net/http"
	"ppo/internal/delivery/http/middleware"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	docs "ppo/internal/delivery/http/docs"
	"ppo/internal/delivery/http/handlers"
	"ppo/internal/services"
)

func NewRouter(
	catSvc services.CategoryService,
	dsSvc services.DatasetService,
	notifSvc services.NotificationService,
	revSvc services.ReviewService,
	userSvc services.UserService,
	subSvc services.SubscriptionService,
	logger *zap.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware(logger))

	docs.SwaggerInfo.BasePath = "/"

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", middleware.WrapHandler(handlers.ListCategories(catSvc)))

		r.Post("/", middleware.WrapHandler(handlers.CreateCategory(catSvc)))

		r.Get("/{id}", middleware.WrapHandler(handlers.GetCategory(catSvc)))

		r.Put("/{id}", middleware.WrapHandler(handlers.UpdateCategory(catSvc)))

		r.Delete("/{id}", middleware.WrapHandler(handlers.DeleteCategory(catSvc)))
	})

	r.Route("/datasets", func(r chi.Router) {
		r.Get("/", middleware.WrapHandler(handlers.ListDatasets(dsSvc)))

		r.Post("/", middleware.WrapHandler(handlers.CreateDataset(dsSvc)))

		r.Get("/{id}", middleware.WrapHandler(handlers.GetDataset(dsSvc)))

		r.Post("/{id}/versions", middleware.WrapHandler(handlers.AddVersion(dsSvc)))

		r.Get("/{id}/versions", middleware.WrapHandler(handlers.ListVersions(dsSvc)))

		r.Get("/{id}/subscribers", middleware.WrapHandler(handlers.ListSubscribersHandler(subSvc)))

		r.Post("/{id}/notifications", middleware.WrapHandler(handlers.NotifySubscribersHandler(notifSvc)))

		r.Get("/{id}/reviews", middleware.WrapHandler(handlers.ListReviewsByDatasetHandler(revSvc)))

		r.Get("/{id}/reviews/summary", middleware.WrapHandler(handlers.GetRatingSummaryHandler(revSvc)))
	})

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", middleware.WrapHandler(handlers.SubscribeHandler(subSvc)))

		r.Delete("/", middleware.WrapHandler(handlers.UnsubscribeHandler(subSvc)))
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/register", middleware.WrapHandler(handlers.RegisterUserHandler(userSvc)))

		r.Post("/authenticate", middleware.WrapHandler(handlers.AuthenticateUserHandler(userSvc)))

		r.Get("/{id}", middleware.WrapHandler(handlers.GetUserHandler(userSvc)))

		r.Put("/{id}", middleware.WrapHandler(handlers.UpdateUserHandler(userSvc)))

		r.Delete("/{id}", middleware.WrapHandler(handlers.DeleteUserHandler(userSvc)))

		r.Get("/{id}/notifications", middleware.WrapHandler(handlers.GetUserNotificationsHandler(notifSvc)))

		r.Get("/{id}/subscriptions", middleware.WrapHandler(handlers.ListSubscriptionsHandler(subSvc)))

		r.Get("/{id}/reviews", middleware.WrapHandler(handlers.ListReviewsByUserHandler(revSvc)))
	})

	r.Route("/reviews", func(r chi.Router) {
		r.Post("/", middleware.WrapHandler(handlers.CreateReviewHandler(revSvc)))

		r.Get("/{id}", middleware.WrapHandler(handlers.GetReviewByIDHandler(revSvc)))

		r.Put("/{id}", middleware.WrapHandler(handlers.UpdateReviewHandler(revSvc)))

		r.Delete("/{id}", middleware.WrapHandler(handlers.DeleteReviewHandler(revSvc)))
	})

	return r
}
