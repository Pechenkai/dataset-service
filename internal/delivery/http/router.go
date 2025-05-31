package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

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
) http.Handler {
	r := chi.NewRouter()

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/categories", func(r chi.Router) {
		r.Get("/", handlers.ListCategories(catSvc))

		r.Post("/", handlers.CreateCategory(catSvc))

		r.Get("/{id}", handlers.GetCategory(catSvc))

		r.Put("/{id}", handlers.UpdateCategory(catSvc))

		r.Delete("/{id}", handlers.DeleteCategory(catSvc))
	})

	r.Route("/datasets", func(r chi.Router) {
		r.Get("/", handlers.ListDatasets(dsSvc))

		r.Post("/", handlers.CreateDataset(dsSvc))

		r.Get("/{id}", handlers.GetDataset(dsSvc))

		r.Post("/{id}/versions", handlers.AddVersion(dsSvc))

		r.Get("/{id}/versions", handlers.ListVersions(dsSvc))

		r.Get("/{id}/subscribers", handlers.ListSubscribersHandler(subSvc))

		r.Post("/{id}/notifications", handlers.NotifySubscribersHandler(notifSvc))

		r.Get("/{id}/reviews", handlers.ListReviewsByDatasetHandler(revSvc))

		r.Get("/{id}/reviews/summary", handlers.GetRatingSummaryHandler(revSvc))
	})

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", handlers.SubscribeHandler(subSvc))

		r.Delete("/", handlers.UnsubscribeHandler(subSvc))
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/register", handlers.RegisterUserHandler(userSvc))

		r.Post("/authenticate", handlers.AuthenticateUserHandler(userSvc))

		r.Get("/{id}", handlers.GetUserHandler(userSvc))

		r.Put("/{id}", handlers.UpdateUserHandler(userSvc))

		r.Delete("/{id}", handlers.DeleteUserHandler(userSvc))

		r.Get("/{id}/notifications", handlers.GetUserNotificationsHandler(notifSvc))

		r.Get("/{id}/subscriptions", handlers.ListSubscriptionsHandler(subSvc))

		r.Get("/{id}/reviews", handlers.ListReviewsByUserHandler(revSvc))
	})

	r.Route("/reviews", func(r chi.Router) {
		r.Post("/", handlers.CreateReviewHandler(revSvc))

		r.Get("/{id}", handlers.GetReviewByIDHandler(revSvc))

		r.Put("/{id}", handlers.UpdateReviewHandler(revSvc))

		r.Delete("/{id}", handlers.DeleteReviewHandler(revSvc))
	})

	return r
}
