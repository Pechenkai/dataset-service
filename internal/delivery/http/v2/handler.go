package v2

import (
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/http/middleware"
	"ppo/internal/services"
)

type HandlerDeps struct {
	Categories    services.CategoryService
	Datasets      services.DatasetService
	Reviews       services.ReviewService
	Notifications services.NotificationService
	Subscriptions services.SubscriptionService
	Users         services.UserService
	Tokens        services.TokenService
	TwoFA         services.TwoFactorService
	Access        services.AccessService
	Summaries     services.DatasetSummaryService
	Facts         services.DatasetFactService
	Logger        *zap.Logger
	TokenTTL      time.Duration
}

type Handler struct {
	categories    services.CategoryService
	datasets      services.DatasetService
	reviews       services.ReviewService
	notifications services.NotificationService
	subscriptions services.SubscriptionService
	users         services.UserService
	tokens        services.TokenService
	twofa         services.TwoFactorService
	access        services.AccessService
	summaries     services.DatasetSummaryService
	facts         services.DatasetFactService
	logger        *zap.Logger
	tokenTTL      time.Duration
}

func NewHandler(deps HandlerDeps) *Handler {
	tokenTTL := deps.TokenTTL
	if tokenTTL <= 0 {
		tokenTTL = 24 * time.Hour
	}
	return &Handler{
		categories:    deps.Categories,
		datasets:      deps.Datasets,
		reviews:       deps.Reviews,
		notifications: deps.Notifications,
		subscriptions: deps.Subscriptions,
		users:         deps.Users,
		tokens:        deps.Tokens,
		twofa:         deps.TwoFA,
		access:        deps.Access,
		summaries:     deps.Summaries,
		facts:         deps.Facts,
		logger:        deps.Logger,
		tokenTTL:      tokenTTL,
	}
}

func RegisterRoutes(r chi.Router, basePath string, deps HandlerDeps) {
	if basePath == "" {
		basePath = "/api/v2"
	}
	h := NewHandler(deps)

	r.Route(basePath, func(api chi.Router) {
		api.Get("/openapi.yaml", h.serveOpenAPISpec)

		protected := api
		optionalAuth := api
		if deps.Tokens != nil && deps.Users != nil {
			protected = api.With(middleware.BearerAuth(deps.Tokens, deps.Users))
			optionalAuth = api.With(middleware.OptionalBearerAuth(deps.Tokens, deps.Users))
		}

		// Categories - опциональная аутентификация для GET
		optionalAuth.Get("/categories", wrap(deps.Logger, h.ListCategories))
		protected.Post("/categories", wrap(deps.Logger, h.CreateCategory))
		optionalAuth.Get("/categories/{categoryId}", wrap(deps.Logger, h.GetCategory))
		protected.Patch("/categories/{categoryId}", wrap(deps.Logger, h.UpdateCategory))
		protected.Delete("/categories/{categoryId}", wrap(deps.Logger, h.DeleteCategory))

		// Datasets - опциональная аутентификация для GET
		optionalAuth.Get("/datasets", wrap(deps.Logger, h.ListDatasets))
		protected.Post("/datasets", wrap(deps.Logger, h.CreateDataset))
		optionalAuth.Get("/datasets/{datasetId}", wrap(deps.Logger, h.GetDataset))
		protected.Patch("/datasets/{datasetId}", wrap(deps.Logger, h.UpdateDataset))
		protected.Delete("/datasets/{datasetId}", wrap(deps.Logger, h.DeleteDataset))
		optionalAuth.Get("/datasets/{datasetId}/versions", wrap(deps.Logger, h.ListDatasetVersions))
		optionalAuth.Get("/datasets/{datasetId}/summary", wrap(deps.Logger, h.GetDatasetSummary))
		optionalAuth.Get("/datasets/{datasetId}/fun-fact", wrap(deps.Logger, h.GetDatasetFunFact))
		protected.Post("/datasets/{datasetId}/versions", wrap(deps.Logger, h.CreateDatasetVersion))
		optionalAuth.Get("/datasets/{datasetId}/versions/{versionId}", wrap(deps.Logger, h.GetDatasetVersion))
		optionalAuth.Get("/datasets/{datasetId}/versions/{versionId}/content", wrap(deps.Logger, h.DownloadDatasetVersion))
		protected.Post("/datasets/{datasetId}/notifications", wrap(deps.Logger, h.NotifyDatasetSubscribers))
		protected.Get("/datasets/{datasetId}/subscribers", wrap(deps.Logger, h.ListDatasetSubscribers))

		protected.Get("/notifications", wrap(deps.Logger, h.ListNotifications))
		protected.Get("/notifications/{notificationId}", wrap(deps.Logger, h.GetNotification))
		protected.Patch("/notifications/{notificationId}", wrap(deps.Logger, h.UpdateNotification))

		// Reviews - GET доступен без аутентификации
		api.Get("/reviews", wrap(deps.Logger, h.ListReviews))
		protected.Post("/reviews", wrap(deps.Logger, h.CreateReview))
		api.Get("/reviews/{reviewId}", wrap(deps.Logger, h.GetReview))
		protected.Patch("/reviews/{reviewId}", wrap(deps.Logger, h.UpdateReview))
		protected.Delete("/reviews/{reviewId}", wrap(deps.Logger, h.DeleteReview))

		protected.Get("/subscriptions", wrap(deps.Logger, h.ListSubscriptions))
		protected.Post("/subscriptions", wrap(deps.Logger, h.CreateSubscription))
		protected.Get("/subscriptions/{subscriptionId}", wrap(deps.Logger, h.GetSubscription))
		protected.Delete("/subscriptions/{subscriptionId}", wrap(deps.Logger, h.DeleteSubscription))

		protected.Get("/users", wrap(deps.Logger, h.ListUsers))
		api.Post("/users", wrap(deps.Logger, h.RegisterUser))
		protected.Get("/users/{userId}", wrap(deps.Logger, h.GetUser))
		protected.Patch("/users/{userId}", wrap(deps.Logger, h.UpdateUser))
		protected.Delete("/users/{userId}", wrap(deps.Logger, h.DeleteUser))
		protected.Get("/users/{userId}/notifications", wrap(deps.Logger, h.ListUserNotifications))
		protected.Get("/users/{userId}/reviews", wrap(deps.Logger, h.ListUserReviews))
		protected.Get("/users/{userId}/subscriptions", wrap(deps.Logger, h.ListUserSubscriptions))

		// Access Requests
		protected.Get("/access-requests", wrap(deps.Logger, h.ListAccessRequests))
		protected.Post("/access-requests", wrap(deps.Logger, h.CreateAccessRequest))
		protected.Get("/access-requests/{requestId}", wrap(deps.Logger, h.GetAccessRequest))
		protected.Patch("/access-requests/{requestId}", wrap(deps.Logger, h.UpdateAccessRequest))

		api.Post("/auth/2fa/challenge", wrap(deps.Logger, h.RequestTwoFA))
		api.Get("/auth/2fa/challenges/{challengeId}/code", wrap(deps.Logger, h.DebugChallengeCode))
		api.Post("/auth/tokens", wrap(deps.Logger, h.IssueToken))
		protected.Post("/auth/tokens/revoke", wrap(deps.Logger, h.RevokeToken))
	})
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func wrap(logger *zap.Logger, h handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			status := middleware.MapErrorToStatus(err)
			writeError(w, status, err)
			if logger != nil {
				logger.Warn("v2 handler error",
					zap.Error(err),
					zap.Int("status", status),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)
			}
		}
	}
}
