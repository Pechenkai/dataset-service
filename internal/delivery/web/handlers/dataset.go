package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

// DatasetHandler отвечает за CRUD операции с Dataset.
type DatasetHandler struct {
	service services.DatasetService
	logger  *zap.Logger
}

// NewDatasetHandler создаёт новый DatasetHandler без парсинга всех шаблонов.
func NewDatasetHandler(svc services.DatasetService, log *zap.Logger) *DatasetHandler {
	return &DatasetHandler{
		service: svc,
		logger:  log,
	}
}

// RegisterRoutes регистрирует маршруты для операций с Dataset.
func (h *DatasetHandler) RegisterRoutes(r chi.Router) {
	r.Get("/datasets", h.List)
	r.Get("/datasets/new", h.NewForm)
	r.Post("/datasets", h.Create)
	r.Get("/datasets/{id}", h.Show)
	// в будущем: редактирование/удаление/версии
}

// List отображает страницу со списком всех доступных (public) датасетов.
// GET /datasets
func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListDatasets(r.Context(), true, nil)
	if err != nil {
		h.logger.Error("ListDatasets failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Подготовка DTO
	datasets := dto.ToDatasetDTOs(list)

	// Парсим только layout.tmpl + dataset_list.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"dataset_list.tmpl",
	))

	// Передаём в шаблон Title и сам список
	data := struct {
		Title    string
		Datasets []*dto.DatasetDTO
	}{
		Title:    "Список датасетов",
		Datasets: datasets,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Show отображает детальную страницу одного датасета (по его ID).
// GET /datasets/{id}
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

	// Преобразуем в DTO
	datasetDTO := dto.ToDatasetDTO(ds)

	// Парсим только layout.tmpl + dataset_show.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"dataset_show.tmpl",
	))

	data := struct {
		Title   string
		Dataset *dto.DatasetDTO
	}{
		Title:   fmt.Sprintf("Датасет #%d", ds.ID),
		Dataset: datasetDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// NewForm отображает форму создания нового датасета.
// GET /datasets/new
func (h *DatasetHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	// DTO для формы (можно расширить, если нужны категории/метаданные и т. д.)
	formDTO := &dto.CreateDatasetForm{}

	// Парсим только layout.tmpl + dataset_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"dataset_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateDatasetForm
	}{
		Title:      "Новый датасет",
		FormAction: "/datasets",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
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

	cmd := services.CreateDatasetCmd{
		// TODO: ActorID брать из сессии/контекста, пока фиксируем 0
		ActorID:     0,
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
