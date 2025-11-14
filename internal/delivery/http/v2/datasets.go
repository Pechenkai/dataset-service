package v2

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/delivery/http/middleware"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListDatasets(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 20, 200)
	if err != nil {
		return err
	}

	q := r.URL.Query()
	ownerID, err := parseUintPtr(q, "owner_id")
	if err != nil {
		return err
	}
	categoryID, err := parseUintPtr(q, "category_id")
	if err != nil {
		return err
	}
	publicFilter, err := parseBoolPtr(q, "is_public")
	if err != nil {
		return err
	}
	search := strings.ToLower(strings.TrimSpace(q.Get("search")))
	minRating, err := parseRatingBound(q, "min_rating")
	if err != nil {
		return err
	}
	maxRating, err := parseRatingBound(q, "max_rating")
	if err != nil {
		return err
	}
	if minRating != nil && maxRating != nil && *minRating > *maxRating {
		return &dto.BadRequestError{Message: "min_rating cannot be greater than max_rating"}
	}

	onlyPublic := publicFilter != nil && *publicFilter
	list, err := h.datasets.ListDatasets(r.Context(), onlyPublic, ownerID)
	if err != nil {
		return err
	}

	type datasetView struct {
		ds      *entities.Dataset
		summary services.RatingSummary
	}

	views := make([]datasetView, 0, len(list))
	for _, ds := range list {
		if categoryID != nil && ds.CategoryID != *categoryID {
			continue
		}
		if publicFilter != nil && ds.IsPublic != *publicFilter {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(ds.Name), search) && !strings.Contains(strings.ToLower(ds.Description), search) {
			continue
		}
		summary, err := h.reviews.GetRatingSummary(r.Context(), ds.ID)
		if err != nil {
			return err
		}
		if minRating != nil && summary.Average < *minRating {
			continue
		}
		if maxRating != nil && summary.Average > *maxRating {
			continue
		}
		views = append(views, datasetView{ds: ds, summary: summary})
	}

	total := len(views)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)

	items := make([]DatasetResponse, 0, end-start)
	for _, view := range views[start:end] {
		items = append(items, h.toDatasetResponse(view.ds, view.summary, nil))
	}

	resp := DatasetsResponse{
		Items: items,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) GetDataset(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	ds, err := h.datasets.GetDataset(r.Context(), id)
	if err != nil {
		return err
	}

	summary, err := h.reviews.GetRatingSummary(r.Context(), id)
	if err != nil {
		return err
	}

	latest, err := h.latestVersion(r.Context(), id)
	if err != nil && !errors.Is(err, services.ErrVersionNotFound) {
		return err
	}

	var latestResp *DatasetVersionResponse
	if latest != nil {
		latestResp, err = h.buildVersionResponse(r.Context(), latest, ds.OwnerID)
		if err != nil {
			return err
		}
	}

	resp := h.toDatasetResponse(ds, summary, latestResp)
	writeJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) latestVersion(ctx context.Context, datasetID uint64) (*entities.DatasetVersion, error) {
	versions, err := h.datasets.ListVersions(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, nil
	}
	sort.SliceStable(versions, func(i, j int) bool {
		return versions[i].UploadDate.Before(versions[j].UploadDate)
	})
	return versions[len(versions)-1], nil
}

func (h *Handler) toDatasetResponse(ds *entities.Dataset, summary services.RatingSummary, latest *DatasetVersionResponse) DatasetResponse {
	resp := DatasetResponse{
		ID:          ds.ID,
		Name:        ds.Name,
		Description: ds.Description,
		CategoryID:  ds.CategoryID,
		OwnerID:     ds.OwnerID,
		IsPublic:    ds.IsPublic,
		CreatedAt:   ds.CreatedAt,
		RatingSummary: DatasetRatingSummary{
			Average: round(summary.Average, 2),
			Count:   summary.Count,
			Synced:  time.Now().UTC(),
		},
	}
	if latest != nil {
		resp.LatestVersion = latest
		if latest.Metadata != nil {
			resp.Metadata = latest.Metadata
		}
	}
	return resp
}

func (h *Handler) CreateDataset(w http.ResponseWriter, r *http.Request) error {
	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return &dto.BadRequestError{Message: "invalid multipart form"}
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return &dto.BadRequestError{Message: "name is required"}
	}
	categoryRaw := firstNonEmpty(r.FormValue("category_id"), r.FormValue("category"))
	catID, err := strconv.ParseUint(categoryRaw, 10, 64)
	if err != nil || catID == 0 {
		return &dto.BadRequestError{Message: "category_id must be a positive integer"}
	}
	description := r.FormValue("description")
	publicRaw := firstNonEmpty(r.FormValue("is_public"), r.FormValue("public"))
	isPublic := false
	if publicRaw != "" {
		val, err := strconv.ParseBool(publicRaw)
		if err != nil {
			return &dto.BadRequestError{Message: "is_public must be boolean"}
		}
		isPublic = val
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return &dto.BadRequestError{Message: "file is required"}
	}
	defer file.Close()

	metaFormat := firstNonEmpty(r.FormValue("meta_format"), r.FormValue("metaFormat"))
	metaTags := collectMultipartTags(r.MultipartForm, "meta_tags", "metaTags")
	metaSizeStr := firstNonEmpty(r.FormValue("meta_size"), r.FormValue("metaSize"))
	var metaSize uint64
	if metaSizeStr != "" {
		metaSize, err = strconv.ParseUint(metaSizeStr, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "meta_size must be numeric"}
		}
	}

	cmd := services.CreateDatasetCmd{
		ActorID:     user.ID,
		Name:        name,
		Description: description,
		CategoryID:  catID,
		IsPublic:    isPublic,
		FileName:    header.Filename,
		MetaFormat:  metaFormat,
		MetaTags:    metaTags,
		MetaSize:    metaSize,
	}
	id, err := h.datasets.CreateDataset(r.Context(), cmd, file, header.Size)
	if err != nil {
		return err
	}

	ds, err := h.datasets.GetDataset(r.Context(), id)
	if err != nil {
		return err
	}
	summary, err := h.reviews.GetRatingSummary(r.Context(), id)
	if err != nil {
		return err
	}
	latest, err := h.latestVersion(r.Context(), id)
	if err != nil && !errors.Is(err, services.ErrVersionNotFound) {
		return err
	}
	var latestResp *DatasetVersionResponse
	if latest != nil {
		latestResp, err = h.buildVersionResponse(r.Context(), latest, ds.OwnerID)
		if err != nil {
			return err
		}
	}

	writeJSON(w, http.StatusCreated, h.toDatasetResponse(ds, summary, latestResp))
	return nil
}

func (h *Handler) UpdateDataset(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		CategoryID  *uint64 `json:"category_id"`
		IsPublic    *bool   `json:"is_public"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.Name == nil && req.Description == nil && req.CategoryID == nil && req.IsPublic == nil {
		return &dto.BadRequestError{Message: "nothing to update"}
	}

	ds, err := h.datasets.GetDataset(r.Context(), id)
	if err != nil {
		return err
	}

	name := ds.Name
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return &dto.BadRequestError{Message: "name cannot be empty"}
		}
		name = *req.Name
	}
	desc := ds.Description
	if req.Description != nil {
		desc = *req.Description
	}
	categoryID := ds.CategoryID
	if req.CategoryID != nil {
		if *req.CategoryID == 0 {
			return &dto.BadRequestError{Message: "category_id must be positive"}
		}
		categoryID = *req.CategoryID
	}
	isPublic := ds.IsPublic
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	cmd := services.UpdateDatasetCmd{
		ID:          id,
		Name:        name,
		Description: desc,
		CategoryID:  categoryID,
		IsPublic:    isPublic,
	}
	if err := h.datasets.UpdateDataset(r.Context(), cmd); err != nil {
		return err
	}

	updated, err := h.datasets.GetDataset(r.Context(), id)
	if err != nil {
		return err
	}
	summary, err := h.reviews.GetRatingSummary(r.Context(), id)
	if err != nil {
		return err
	}
	latest, err := h.latestVersion(r.Context(), id)
	if err != nil && !errors.Is(err, services.ErrVersionNotFound) {
		return err
	}
	var latestResp *DatasetVersionResponse
	if latest != nil {
		latestResp, err = h.buildVersionResponse(r.Context(), latest, updated.OwnerID)
		if err != nil {
			return err
		}
	}

	writeJSON(w, http.StatusOK, h.toDatasetResponse(updated, summary, latestResp))
	return nil
}

func (h *Handler) DeleteDataset(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	if err := h.datasets.DeleteDataset(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) ListDatasetVersions(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	limit, offset, err := parsePagination(r, 20, 100)
	if err != nil {
		return err
	}

	ds, err := h.datasets.GetDataset(r.Context(), datasetID)
	if err != nil {
		return err
	}

	versions, err := h.datasets.ListVersions(r.Context(), datasetID)
	if err != nil {
		return err
	}

	total := len(versions)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)

	items := make([]DatasetVersionResponse, 0, end-start)
	for _, ver := range versions[start:end] {
		resp, err := h.buildVersionResponse(r.Context(), ver, ds.OwnerID)
		if err != nil {
			return err
		}
		items = append(items, *resp)
	}

	writeJSON(w, http.StatusOK, DatasetVersionsResponse{
		Items: items,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
	return nil
}

func (h *Handler) CreateDatasetVersion(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}
	ds, err := h.datasets.GetDataset(r.Context(), datasetID)
	if err != nil {
		return err
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return &dto.BadRequestError{Message: "invalid multipart form"}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return &dto.BadRequestError{Message: "file is required"}
	}
	defer file.Close()

	cmd := services.AddVersionCmd{
		ActorID:    user.ID,
		DatasetID:  datasetID,
		ChangeLog:  r.FormValue("change_log"),
		FileName:   header.Filename,
		MetaFormat: firstNonEmpty(r.FormValue("meta_format"), r.FormValue("metaFormat")),
		MetaTags:   collectMultipartTags(r.MultipartForm, "meta_tags", "metaTags"),
	}
	if metaSize := firstNonEmpty(r.FormValue("meta_size"), r.FormValue("metaSize")); metaSize != "" {
		if sz, err := strconv.ParseUint(metaSize, 10, 64); err == nil {
			cmd.MetaSize = sz
		}
	}

	versionID, err := h.datasets.AddDatasetVersion(r.Context(), cmd, file, header.Size)
	if err != nil {
		return err
	}

	version, err := h.datasets.GetVersion(r.Context(), versionID)
	if err != nil {
		return err
	}

	resp, err := h.buildVersionResponse(r.Context(), version, ds.OwnerID)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusCreated, resp)
	return nil
}

func (h *Handler) GetDatasetVersion(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	versionID, err := parseIDParam(chi.URLParam(r, "versionId"))
	if err != nil {
		return err
	}

	ds, err := h.datasets.GetDataset(r.Context(), datasetID)
	if err != nil {
		return err
	}

	version, err := h.datasets.GetVersion(r.Context(), versionID)
	if err != nil {
		return err
	}
	if version.DatasetID != datasetID {
		return services.ErrVersionNotFound
	}

	resp, err := h.buildVersionResponse(r.Context(), version, ds.OwnerID)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) NotifyDatasetSubscribers(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	var req struct {
		Message string `json:"message"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Message) == "" {
		return &dto.BadRequestError{Message: "message is required"}
	}

	cmd := services.NotifySubscribersCmd{
		DatasetID: datasetID,
		Message:   req.Message,
	}
	count, err := h.notifications.NotifySubscribers(r.Context(), cmd)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusAccepted, map[string]int{"sent": count})
	return nil
}

func (h *Handler) ListDatasetSubscribers(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	limit, offset, err := parsePagination(r, 50, 500)
	if err != nil {
		return err
	}

	subscribers, err := h.subscriptions.ListSubscribers(r.Context(), datasetID)
	if err != nil {
		return err
	}

	total := len(subscribers)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)

	items := make([]SubscriberResponse, 0, end-start)
	for i := start; i < end; i++ {
		items = append(items, SubscriberResponse{UserID: subscribers[i].UserID})
	}

	writeJSON(w, http.StatusOK, SubscribersResponse{
		Items: items,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
	return nil
}

func (h *Handler) buildVersionResponse(ctx context.Context, ver *entities.DatasetVersion, ownerID uint64) (*DatasetVersionResponse, error) {
	if ver == nil {
		return nil, nil
	}
	md, err := h.datasets.GetVersionMetadata(ctx, ver.ID)
	if err != nil {
		return nil, err
	}
	resp := &DatasetVersionResponse{
		ID:         ver.ID,
		DatasetID:  ver.DatasetID,
		Number:     ver.Number,
		ChangeLog:  ver.ChangeLog,
		FileURL:    ver.Filepath,
		UploadAt:   ver.UploadDate,
		UploadedBy: ownerID,
		Metadata:   toDatasetMetadata(md),
	}
	return resp, nil
}

func parseUintPtr(q url.Values, key string) (*uint64, error) {
	value := q.Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, &dto.BadRequestError{Message: key + " must be a positive integer"}
	}
	return &parsed, nil
}

func parseBoolPtr(q url.Values, key string) (*bool, error) {
	value := q.Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, &dto.BadRequestError{Message: key + " must be boolean"}
	}
	return &parsed, nil
}

func parseRatingBound(q url.Values, key string) (*float64, error) {
	value := q.Get(key)
	if value == "" {
		return nil, nil
	}
	intVal, err := strconv.Atoi(value)
	if err != nil || intVal < 1 || intVal > 5 {
		return nil, &dto.BadRequestError{Message: key + " must be between 1 and 5"}
	}
	f := float64(intVal)
	return &f, nil
}

func round(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}

func toDatasetMetadata(md *entities.Metadata) *DatasetMetadata {
	if md == nil {
		return nil
	}
	return &DatasetMetadata{
		Format: md.Format,
		Tags:   splitTags(md.Tags),
		Size:   md.Size,
	}
}

func splitTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func currentUserOrError(ctx context.Context) (*entities.User, error) {
	user := middleware.CurrentUser(ctx)
	if user == nil {
		return nil, services.ErrInvalidCredentials
	}
	return user, nil
}
