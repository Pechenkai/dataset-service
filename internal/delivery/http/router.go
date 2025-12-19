package delivhttp

import (
	"go.uber.org/zap"
	"net/http"
	"ppo/internal/delivery/http/middleware"
	"ppo/internal/metrics"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	docs "ppo/internal/delivery/http/docs"
	"ppo/internal/delivery/http/handlers"
	v2 "ppo/internal/delivery/http/v2"
	"ppo/internal/services"
)

func NewRouter(
	catSvc services.CategoryService,
	dsSvc services.DatasetService,
	notifSvc services.NotificationService,
	revSvc services.ReviewService,
	userSvc services.UserService,
	subSvc services.SubscriptionService,
	reqSvc services.AccessService,
	tokenSvc services.TokenService,
	tokenTTL time.Duration,
	twofaSvc services.TwoFactorService,
	logger *zap.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(metrics.Middleware)
	r.Use(middleware.LoggingMiddleware(logger))

	docs.SwaggerInfo.BasePath = "/"

	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.Handler().ServeHTTP(w, r)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	registerLegacyRoutes(r, catSvc, dsSvc, notifSvc, revSvc, userSvc, subSvc)

	r.Route("/api/v1", func(api chi.Router) {
		registerLegacyRoutes(api, catSvc, dsSvc, notifSvc, revSvc, userSvc, subSvc)
	})

	v2.RegisterRoutes(r, "/api/v2", v2.HandlerDeps{
		Categories:    catSvc,
		Datasets:      dsSvc,
		Reviews:       revSvc,
		Notifications: notifSvc,
		Subscriptions: subSvc,
		Users:         userSvc,
		Tokens:        tokenSvc,
		TwoFA:         twofaSvc,
		Access:        reqSvc,
		Logger:        logger,
		TokenTTL:      tokenTTL,
	})

	// Swagger UI для v2 API
	r.Get("/swagger/v2", v2.SwaggerUIHandler("/api/v2/openapi.yaml"))
	r.Get("/api/v2/swagger", v2.SwaggerUIHandler("/api/v2/openapi.yaml"))

	registerSPARoutes(r)

	return r
}

func registerLegacyRoutes(
	r chi.Router,
	catSvc services.CategoryService,
	dsSvc services.DatasetService,
	notifSvc services.NotificationService,
	revSvc services.ReviewService,
	userSvc services.UserService,
	subSvc services.SubscriptionService,
) {
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
}

func registerSPARoutes(r chi.Router) {
	const spaRoot = "static/app"
	fsRoot := http.Dir(spaRoot)

	// Главная SPA страница
	r.Get("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, spaRoot+"/index.html")
	})

	// Отдача ассетов + fallback на index.html для роутера
	r.Handle("/app/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/app")
		path = strings.TrimPrefix(path, "/")
		if path == "" {
			http.ServeFile(w, r, spaRoot+"/index.html")
			return
		}

		f, err := fsRoot.Open(path)
		if err == nil {
			defer f.Close()
			if info, _ := f.Stat(); info != nil && !info.IsDir() {
				// fs.File может не реализовывать io.ReadSeeker; используем ServeFile с прямым путём
				http.ServeFile(w, r, spaRoot+"/"+path)
				return
			}
		}

		http.ServeFile(w, r, spaRoot+"/index.html")
	}))
}
