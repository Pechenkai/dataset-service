package handlers

import (
	"errors"
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
	notificationService services.NotificationService
	logger              *zap.Logger
}

func NewDatasetHandler(svc services.DatasetService, rewsvc services.ReviewService, usersvc services.UserService,
	catsvc services.CategoryService, subsvc services.SubscriptionService, notifsvc services.NotificationService, log *zap.Logger) *DatasetHandler {
	return &DatasetHandler{
		service:             svc,
		logger:              log,
		reviewService:       rewsvc,
		userService:         usersvc,
		notificationService: notifsvc,
		categoryService:     catsvc,
		subscriptionService: subsvc,
	}
}

func (h *DatasetHandler) RegisterRoutes(r chi.Router) {
	r.Get("/datasets", h.List)
	r.With(middleware.RequireRole("user", "admin")).Get("/datasets/new", h.NewForm)
	r.Post("/datasets", h.Create)
	r.Get("/datasets/{id}", h.Show)
	r.With(middleware.RequireRole("user", "admin")).Get("/datasets/{id}/edit", h.EditForm)
	r.With(middleware.RequireRole("user", "admin")).Post("/datasets/{id}", h.Update)
	r.With(middleware.RequireRole("user", "admin")).Get("/my-datasets", h.MyList)
	r.With(middleware.RequireRole("user", "admin")).Post("/datasets/{id}/delete", h.Delete)
	r.With(middleware.RequireRole("user", "admin")).Get(
		"/datasets/{id}/versions/new", h.AddVersionForm)
	r.With(middleware.RequireRole("user", "admin")).Post(
		"/datasets/{id}/versions", h.AddVersion)
	r.With(middleware.RequireRole("user", "admin")).Get(
		"/datasets/{id}/versions", h.ListVersions)
	r.With(middleware.RequireRole("user", "admin")).Get(
		"/datasets/{id}/versions/{vid}/download", h.DownloadVersion)
}

func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListDatasets(r.Context(), false, nil)
	if err != nil {
		h.logger.Error("ListDatasets failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dtos := dto.ToDatasetDTOs(list)

	usernameMap := make(map[uint64]string, len(list))
	for _, ds := range list {
		if _, ok := usernameMap[ds.OwnerID]; !ok {
			user, err := h.userService.GetUserByID(r.Context(), ds.OwnerID)
			if err == nil && user != nil {
				usernameMap[ds.OwnerID] = user.Username
			} else {
				usernameMap[ds.OwnerID] = "неизвестный"
			}
		}
	}
	for i, ds := range list {
		dtos[i].OwnerName = usernameMap[ds.OwnerID]
	}

	currentUID, _ := middleware.FromContext(r.Context())
	for i, ds := range list {
		if ds.IsPublic {
			subscribed, _ := h.subscriptionService.IsSubscribed(r.Context(), currentUID, ds.ID)
			dtos[i].IsSubscribed = subscribed
			url, err := h.service.GetDownloadURL(r.Context(), ds.ID)
			if err == nil {
				dtos[i].DownloadURL = url
			}
		}
	}

	data := struct {
		Role     string
		UserID   uint64
		Datasets []*dto.DatasetDTO
	}{
		Role:     func() string { _, role := middleware.FromContext(r.Context()); return role }(),
		UserID:   func() uint64 { uid, _ := middleware.FromContext(r.Context()); return uid }(),
		Datasets: dtos,
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
			if _, ok := usernameMap[rev.UserID]; !ok {
				user, e := h.userService.GetUserByID(r.Context(), rev.UserID)
				if e == nil && user != nil {
					usernameMap[rev.UserID] = user.Username
				} else {
					usernameMap[rev.UserID] = "неизвестный"
				}
			}
		}
		dtoReviews := dto.ToReviewDTOs(rawReviews)
		for i, rev := range rawReviews {
			dtoReviews[i].AuthorUsername = usernameMap[rev.UserID]
		}
		dtoDS.Reviews = dtoReviews
		dtoDS.HasReviews = len(dtoReviews) > 0
	}

	if dtoDS.IsPublic {
		sub, _ := h.subscriptionService.IsSubscribed(r.Context(), currentUID, ds.ID)
		dtoDS.IsSubscribed = sub

		url, err := h.service.GetDownloadURL(r.Context(), ds.ID)
		if err == nil {
			dtoDS.DownloadURL = url
		}
	}

	url, err := h.service.GetDownloadURL(r.Context(), ds.ID)
	if err == nil {
		dtoDS.DownloadURL = url
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
		IsNew      bool
		Form       *dto.CreateDatasetForm
	}{
		Title:      "Новый датасет",
		Role:       currentRole,
		UserID:     currentUID,
		IsNew:      true,
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

func (h *DatasetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	datasetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	currentUID, currentRole := middleware.FromContext(r.Context())
	ds, err := h.service.GetDataset(r.Context(), datasetID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if currentRole != "admin" && ds.OwnerID != currentUID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.service.DeleteDataset(r.Context(), datasetID); err != nil {
		h.logger.Error("DeleteDataset failed", zap.Error(err), zap.Uint64("dataset_id", datasetID))
		if errors.Is(err, services.ErrDatasetNotFound) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/my-datasets", http.StatusSeeOther)
}

func (h *DatasetHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ds, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if currentRole != "admin" && ds.OwnerID != currentUID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	form := &dto.UpdateDatasetForm{
		ID:          ds.ID,
		Name:        ds.Name,
		Description: ds.Description,
		CategoryID:  ds.CategoryID,
		IsPublic:    ds.IsPublic,
	}

	cats, _ := h.categoryService.ListCategories(r.Context())
	catDTOs := dto.ToCategoryDTOs(cats)

	tpl := template.Must(template.ParseFS(templates.TemplatesFS,
		"layout.tmpl", "dataset_form.tmpl"))
	data := struct {
		Title      string
		Role       string
		UserID     uint64
		IsNew      bool
		FormAction string
		Form       *dto.UpdateDatasetForm
		Categories []*dto.CategoryDTO
	}{
		Title:      fmt.Sprintf("Редактировать датасет #%d", id),
		Role:       currentRole,
		UserID:     currentUID,
		IsNew:      false,
		FormAction: fmt.Sprintf("/datasets/%d", id),
		Form:       form,
		Categories: catDTOs,
	}
	tpl.ExecuteTemplate(w, "layout.tmpl", data)
}

func (h *DatasetHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ds, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if currentRole != "admin" && ds.OwnerID != currentUID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	catID, _ := strconv.ParseUint(r.FormValue("category_id"), 10, 64)
	cmd := services.UpdateDatasetCmd{
		ID:          id,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		CategoryID:  catID,
		IsPublic:    r.FormValue("is_public") == "on",
	}
	if err := h.service.UpdateDataset(r.Context(), cmd); err != nil {
		h.logger.Error("UpdateDataset failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/datasets/%d", id), http.StatusSeeOther)
}

func (h *DatasetHandler) AddVersionForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ds, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if currentRole != "admin" && ds.OwnerID != currentUID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"version_form.tmpl",
	))
	data := struct {
		Title      string
		Role       string
		UserID     uint64
		DatasetID  uint64
		FormAction string
	}{
		Title:      fmt.Sprintf("Добавить версию датасета #%d", id),
		Role:       currentRole,
		UserID:     currentUID,
		DatasetID:  id,
		FormAction: fmt.Sprintf("/datasets/%d/versions", id),
	}
	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("version form template error", zap.Error(err))
	}
}

func (h *DatasetHandler) AddVersion(w http.ResponseWriter, r *http.Request) {
	currentUID, _ := middleware.FromContext(r.Context())
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ds, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if ds.OwnerID != currentUID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	const maxMemory = 10 << 20
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Bad Request: no file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	size := header.Size
	changeLog := r.FormValue("changelog")

	cmd := services.AddVersionCmd{
		DatasetID: id,
		FileName:  header.Filename,
		ChangeLog: changeLog,
	}
	_, err = h.service.AddDatasetVersion(r.Context(), cmd, file, size)
	if err != nil {
		h.logger.Error("AddVersion failed", zap.Error(err), zap.Uint64("dataset", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg := fmt.Sprintf("Для датасета '%s' вышла новая версия", ds.Name)
	notifCmd := services.NotifySubscribersCmd{
		DatasetID: id,
		Message:   msg,
	}
	if _, err := h.notificationService.NotifySubscribers(r.Context(), notifCmd); err != nil {
		h.logger.Error("NotifySubscribers failed", zap.Error(err))
	}

	http.Redirect(w, r, fmt.Sprintf("/datasets/%d", id), http.StatusSeeOther)
}

func (h *DatasetHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	dsID, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	versions, err := h.service.ListVersions(r.Context(), dsID)
	if err != nil {
		h.logger.Error("ListVersions failed", zap.Error(err), zap.Uint64("datasetID", dsID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	dtoVers := dto.ToVersionDTOs(versions)
	for i, v := range dtoVers {
		url, err := h.service.GetVersionDownloadURL(r.Context(), v.ID)
		if err == nil {
			dtoVers[i].DownloadURL = url
		}
	}
	currentUID, currentRole := middleware.FromContext(r.Context())
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"version_list.tmpl",
	))
	data := struct {
		Title     string
		Role      string
		UserID    uint64
		DatasetID uint64
		Versions  []*dto.VersionDTO
	}{
		Title:     fmt.Sprintf("Версии датасета #%d", dsID),
		Role:      currentRole,
		UserID:    currentUID,
		DatasetID: dsID,
		Versions:  dtoVers,
	}
	tpl.ExecuteTemplate(w, "layout.tmpl", data)
}

func (h *DatasetHandler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	vid, _ := strconv.ParseUint(chi.URLParam(r, "vid"), 10, 64)
	url, err := h.service.GetVersionDownloadURL(r.Context(), vid)
	if err != nil {
		h.logger.Error("DownloadVersion failed", zap.Error(err), zap.Uint64("versionID", vid))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}
