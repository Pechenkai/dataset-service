package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"ppo/internal/delivery/web/middleware"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

type DatasetHandler struct {
	service             services.DatasetService
	reviewService       services.ReviewService
	userService         services.UserService
	categoryService     services.CategoryService
	subscriptionService services.SubscriptionService
	logger              *zap.Logger
}

func NewDatasetHandler(svc services.DatasetService, rewsvc services.ReviewService, usersvc services.UserService,
	catsvc services.CategoryService, subsvc services.SubscriptionService, log *zap.Logger) *DatasetHandler {
	return &DatasetHandler{
		service:             svc,
		logger:              log,
		reviewService:       rewsvc,
		userService:         usersvc,
		categoryService:     catsvc,
		subscriptionService: subsvc,
	}
}

func (h *DatasetHandler) RegisterRoutes(r chi.Router) {
	r.Get("/datasets", h.List)
	r.With(middleware.RequireRole("user", "admin")).Get("/datasets/new", h.NewForm)
	r.Post("/datasets", h.Create)
	r.Get("/datasets/{id}", h.Show)
	r.With(middleware.RequireRole("user", "admin")).Get("/my-datasets", h.MyList)
}

func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListDatasets(r.Context(), true, nil)
	if err != nil {
		h.logger.Error("ListDatasets failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		Role     string
		UserID   uint64
		Datasets []*dto.DatasetDTO
	}{
		Role:     func() string { _, role := middleware.FromContext(r.Context()); return role }(),
		UserID:   func() uint64 { uid, _ := middleware.FromContext(r.Context()); return uid }(),
		Datasets: dto.ToDatasetDTOs(list),
	}
	tpl := template.Must(template.ParseFS(templates.TemplatesFS, "layout.tmpl", "dataset_list.tmpl"))
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *DatasetHandler) Show(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ds, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetDataset failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	currentUID, currentRole := middleware.FromContext(r.Context())

	dtoDS := dto.ToDatasetDTO(ds)

	owner, err := h.userService.GetUserByID(r.Context(), ds.OwnerID) // Предположим, у вас есть userService
	if err == nil && owner != nil {
		dtoDS.OwnerName = owner.Username
	}

	cat, err := h.categoryService.GetCategoryByID(r.Context(), ds.CategoryID)
	if err == nil && cat != nil {
		dtoDS.CategoryName = cat.Name
	}

	rawReviews, err := h.reviewService.ListByDataset(r.Context(), id)
	if err == nil {
		usernameMap := make(map[uint64]string)
		for _, rev := range rawReviews {
			if _, seen := usernameMap[rev.UserID]; !seen {
				user, e := h.userService.GetUserByID(r.Context(), rev.UserID)
				if e == nil && user != nil {
					usernameMap[rev.UserID] = user.Username
				} else {
					usernameMap[rev.UserID] = "неизвестный"
				}
			}
		}
		dtoReviews := dto.ToReviewDTOs(rawReviews)
		dtoDS.Reviews = dtoReviews
		dtoDS.HasReviews = len(dtoReviews) > 0
	}

	data := struct {
		Title   string
		Dataset *dto.DatasetDTO
		User    uint64
		Role    string
	}{
		Title:   fmt.Sprintf("Датасет #%d", ds.ID),
		Dataset: dtoDS,
		User:    currentUID,
		Role:    currentRole,
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"dataset_show.tmpl",
	))
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *DatasetHandler) MyList(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	list, err := h.service.ListDatasets(r.Context(), false, &currentUID)
	if err != nil {
		h.logger.Error("ListDatasets(owner) failed", zap.Error(err), zap.Uint64("ownerID", currentUID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Role     string
		UserID   uint64
		Datasets []*dto.DatasetDTO
	}{
		Role:     currentRole,
		UserID:   currentUID,
		Datasets: dto.ToDatasetDTOs(list),
	}
	tpl := template.Must(template.ParseFS(templates.TemplatesFS, "layout.tmpl", "my_dataset_list.tmpl"))
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *DatasetHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	formDTO := &dto.CreateDatasetForm{}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"dataset_form.tmpl",
	))

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		FormAction string
		Form       *dto.CreateDatasetForm
	}{
		Title:      "Новый датасет",
		Role:       currentRole,
		UserID:     currentUID,
		FormAction: "/datasets",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *DatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 10 << 20 // 10 МБ
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		h.logger.Warn("ParseMultipartForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	categoryID64, err := strconv.ParseUint(r.FormValue("category_id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid category_id", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	isPublic := r.FormValue("is_public") == "on"
	metaFormat := r.FormValue("meta_format")
	metaTags := r.FormValue("meta_tags")
	metaSize64, _ := strconv.ParseUint(r.FormValue("meta_size"), 10, 64)

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("FormFile error", zap.Error(err))
		http.Error(w, "Bad Request: no file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	size := header.Size
	fileName := header.Filename

	currentUID, _ := middleware.FromContext(r.Context())

	cmd := services.CreateDatasetCmd{
		ActorID:     currentUID,
		Name:        name,
		Description: description,
		CategoryID:  categoryID64,
		FileName:    fileName,
		IsPublic:    isPublic,
		MetaFormat:  metaFormat,
		MetaTags:    metaTags,
		MetaSize:    metaSize64,
	}

	id, err := h.service.CreateDataset(r.Context(), cmd, file, size)
	if err != nil {
		h.logger.Error("CreateDataset failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/datasets/%d", id), http.StatusSeeOther)
}
