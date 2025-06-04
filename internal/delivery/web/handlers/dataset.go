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

// DatasetHandler отвечает за CRUD операции с Dataset.
type DatasetHandler struct {
	service       services.DatasetService
	reviewService services.ReviewService
	logger        *zap.Logger
}

// NewDatasetHandler создаёт новый DatasetHandler без парсинга всех шаблонов.
func NewDatasetHandler(svc services.DatasetService, rewsvc services.ReviewService, log *zap.Logger) *DatasetHandler {
	return &DatasetHandler{
		service:       svc,
		logger:        log,
		reviewService: rewsvc,
	}
}

// RegisterRoutes регистрирует маршруты для операций с Dataset.
func (h *DatasetHandler) RegisterRoutes(r chi.Router) {
	r.Get("/datasets", h.List)
	r.With(middleware.RequireRole("user", "admin")).Get("/datasets/new", h.NewForm)
	r.Post("/datasets", h.Create)
	r.Get("/datasets/{id}", h.Show)
	r.With(middleware.RequireRole("user", "admin")).Get("/my-datasets", h.MyList)
	// в будущем: редактирование/удаление/версии
}

// List отображает страницу со списком всех доступных (public) датасетов.
// GET /datasets
func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	// onlyPublic = true, ownerID = nil
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
		// title можно не передавать, layout возьмёт из блока title
		Role:     func() string { _, role := middleware.FromContext(r.Context()); return role }(),
		UserID:   func() uint64 { uid, _ := middleware.FromContext(r.Context()); return uid }(),
		Datasets: dto.ToDatasetDTOs(list),
	}
	tpl := template.Must(template.ParseFS(templates.TemplatesFS, "layout.tmpl", "dataset_list.tmpl"))
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Show отображает детальную страницу одного датасета (по его ID).
// GET /datasets/{id}
func (h *DatasetHandler) Show(w http.ResponseWriter, r *http.Request) {
	// 1) Извлечь ID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 2) Получить датасет (включая метаданные, ownerID, isPublic и т.д.)
	ds, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetDataset failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	// 3) Проверка: если это приватный датасет (ds.IsPublic == false) и пользователь не владелец и не админ → 403
	currentUID, currentRole := middleware.FromContext(r.Context())
	if !ds.IsPublic && !(currentRole == "admin" || currentUID == ds.OwnerID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 4) Получить список отзывов для этого датасета
	reviews, err := h.reviewService.ListByDataset(r.Context(), id)
	if err != nil {
		h.logger.Error("ListByDataset failed", zap.Error(err), zap.Uint64("datasetID", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5) Запрос на формы создания нового отзыва:
	canReview := (currentRole == "user" || currentRole == "admin") // или нужно дополнительно проверить, что юзер не автор датасета и что он ещё не оставлял отзыв – по требованию

	// 6) Подготовка DTO
	dsDTO := dto.ToDatasetDTO(ds)
	revDTOs := dto.ToReviewDTOs(reviews) // конвертация списка reviews в DTO

	data := struct {
		Role      string
		UserID    uint64
		Dataset   *dto.DatasetDTO
		Reviews   []*dto.ReviewDTO
		CanReview bool
		// Для формы отзыва:
		ReviewAction string
	}{
		Role:         currentRole,
		UserID:       currentUID,
		Dataset:      dsDTO,
		Reviews:      revDTOs,
		CanReview:    canReview,
		ReviewAction: fmt.Sprintf("/reviews"), // форма POST /reviews с hidden dataset_id
	}

	tpl := template.Must(template.ParseFS(templates.TemplatesFS, "layout.tmpl", "dataset_show.tmpl"))
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *DatasetHandler) MyList(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	// currentRole гарантированно "user" или "admin"

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

// NewForm отображает форму создания нового датасета.
// GET /datasets/new
func (h *DatasetHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	// 1) Получаем из контекста роль и UID
	currentUID, currentRole := middleware.FromContext(r.Context())

	// 2) Создаем пустой DTO для формы
	formDTO := &dto.CreateDatasetForm{}

	// 3) Парсим оба шаблона: layout + dataset_form
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"dataset_form.tmpl",
	))

	// 4) Составляем структуру, которую передадим в шаблон
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

	// 5) Пишем результат в HTTP
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Create обрабатывает POST /datasets: создаёт новый датасет + загружает файл и метаданные.
func (h *DatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 10 << 20 // 10 МБ
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		h.logger.Warn("ParseMultipartForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Читаем поля формы
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
