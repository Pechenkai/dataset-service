package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListDatasets(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 20, 200)
	if err != nil {
		return err
	}
	params, err := parseDatasetListQuery(r.URL.Query())
	if err != nil {
		return err
	}

	actor := currentUser(r.Context())
	visibility, err := resolveDatasetVisibility(actor, params.OwnerID, params.Public)
	if err != nil {
		return err
	}

	list, err := h.datasets.ListDatasets(r.Context(), visibility.onlyPublic, visibility.ownerFilter)
	if err != nil {
		return err
	}

	views, err := h.collectDatasetViews(r.Context(), actor, list, params, visibility.publicFilter)
	if err != nil {
		return err
	}

	items := paginateDatasets(views, limit, offset, h)

	writeJSON(w, http.StatusOK, DatasetsResponse{
		Items: items,
		Meta: PaginationMeta{
			Total:  len(views),
			Limit:  limit,
			Offset: offset,
		},
	})
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
	if err := ensureDatasetReadable(currentUser(r.Context()), ds.OwnerID, ds.IsPublic); err != nil {
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
		latestResp, err = h.buildVersionResponse(r.Context(), latest)
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
		UpdatedAt:   nil, // Поле опциональное в спецификации, в БД нет updated_at
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
	cmd, file, header, err := parseCreateDatasetRequest(r, user.ID)
	if err != nil {
		return err
	}
	defer file.Close()

	id, err := h.datasets.CreateDataset(r.Context(), cmd, file, header.Size)
	if err != nil {
		return err
	}

	resp, err := h.fullDatasetResponse(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, resp)
	return nil
}

func (h *Handler) UpdateDataset(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}

	req, err := decodeUpdateDatasetRequest(r)
	if err != nil {
		return err
	}

	ds, err := h.datasets.GetDataset(r.Context(), id)
	if err != nil {
		return err
	}
	if err := ensureDatasetOwnerOrAdmin(currentUser(r.Context()), ds.OwnerID); err != nil {
		return err
	}

	cmd, err := buildUpdateDatasetCmd(id, ds, req)
	if err != nil {
		return err
	}
	if err := h.datasets.UpdateDataset(r.Context(), cmd); err != nil {
		return err
	}

	resp, err := h.fullDatasetResponse(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) DeleteDataset(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	ds, err := h.datasets.GetDataset(r.Context(), id)
	if err != nil {
		return err
	}
	if err := ensureDatasetOwnerOrAdmin(currentUser(r.Context()), ds.OwnerID); err != nil {
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
	if err := ensureDatasetReadable(currentUser(r.Context()), ds.OwnerID, ds.IsPublic); err != nil {
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
		resp, err := h.buildVersionResponse(r.Context(), ver)
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
	if err := ensureDatasetOwnerOrAdmin(user, ds.OwnerID); err != nil {
		return err
	}

	cmd, file, header, err := parseDatasetVersionRequest(r, user.ID, datasetID)
	if err != nil {
		return err
	}
	defer file.Close()

	versionID, err := h.datasets.AddDatasetVersion(r.Context(), cmd, file, header.Size)
	if err != nil {
		return err
	}

	version, err := h.datasets.GetVersion(r.Context(), versionID)
	if err != nil {
		return err
	}

	resp, err := h.buildVersionResponse(r.Context(), version)
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

	// Проверяем права доступа к датасету
	if err := ensureDatasetReadable(currentUser(r.Context()), ds.OwnerID, ds.IsPublic); err != nil {
		return err
	}

	version, err := h.datasets.GetVersion(r.Context(), versionID)
	if err != nil {
		return err
	}
	if version.DatasetID != datasetID {
		return services.ErrVersionNotFound
	}

	resp, err := h.buildVersionResponse(r.Context(), version)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, resp)
	return nil
}

//func (h *Handler) DownloadDatasetVersion(w http.ResponseWriter, r *http.Request) error {
//	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
//	if err != nil {
//		return err
//	}
//	versionID, err := parseIDParam(chi.URLParam(r, "versionId"))
//	if err != nil {
//		return err
//	}
//
//	ds, err := h.datasets.GetDataset(r.Context(), datasetID)
//	if err != nil {
//		return err
//	}
//	if err := ensureDatasetReadable(currentUser(r.Context()), ds.OwnerID, ds.IsPublic); err != nil {
//		return err
//	}
//
//	version, err := h.datasets.GetVersion(r.Context(), versionID)
//	if err != nil {
//		return err
//	}
//	if version.DatasetID != datasetID {
//		return services.ErrVersionNotFound
//	}
//
//	//url, err := h.datasets.GetVersionDownloadURL(r.Context(), versionID)
//	//if err != nil {
//	//	return err
//	//}
//	//if url == "" {
//	//	return services.ErrVersionNotFound
//	//}
//	//w.Header().Set("Location", url)
//	//w.WriteHeader(http.StatusFound)
//	//return nil
//
//	objectKey, err := h.datasets.GetVersionObjectKey(r.Context(), versionID)
//	if err != nil {
//		return err
//	}
//
//	obj, err := h.storage.GetObject(r.Context(), objectKey) // обёртка над MinIO/S3
//	if err != nil {
//		if errors.Is(err, storage.ErrNotFound) {
//			return services.ErrVersionNotFound
//		}
//		return err
//	}
//	defer obj.Body.Close()
//
//	w.Header().Set("Content-Type", "application/octet-stream")
//	w.Header().Set("Content-Disposition", `attachment; filename="`+version.Filename+`"`)
//
//	if _, err := io.Copy(w, obj.Body); err != nil {
//		// логируем, но клиенту уже отправили тело
//		return nil
//	}
//	return nil
//}

func (h *Handler) DownloadDatasetVersion(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	versionID, err := parseIDParam(chi.URLParam(r, "versionId"))
	if err != nil {
		return err
	}

	version, err := h.loadReadableVersion(r.Context(), currentUser(r.Context()), datasetID, versionID)
	if err != nil {
		return err
	}

	url, err := h.datasets.GetVersionDownloadURL(r.Context(), versionID)
	if err != nil {
		return err
	}
	if url == "" {
		return services.ErrVersionNotFound
	}

	return streamDatasetVersion(r.Context(), w, url, version.Filepath)
}

func (h *Handler) NotifyDatasetSubscribers(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	user := currentUser(r.Context())
	ds, err := h.datasets.GetDataset(r.Context(), datasetID)
	if err != nil {
		return err
	}
	if err := ensureDatasetOwnerOrAdmin(user, ds.OwnerID); err != nil {
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

	ownerNotif, err := h.notifications.NotifyUser(r.Context(), user.ID, datasetID, req.Message)
	if err != nil {
		return err
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v2/notifications/%d", ownerNotif.ID))
	w.Header().Set("X-Notifications-Count", strconv.Itoa(count))
	writeJSON(w, http.StatusCreated, toNotificationResponse(ownerNotif))
	return nil
}

func (h *Handler) ListDatasetSubscribers(w http.ResponseWriter, r *http.Request) error {
	datasetID, err := parseIDParam(chi.URLParam(r, "datasetId"))
	if err != nil {
		return err
	}
	ds, err := h.datasets.GetDataset(r.Context(), datasetID)
	if err != nil {
		return err
	}
	if err := ensureDatasetOwnerOrAdmin(currentUser(r.Context()), ds.OwnerID); err != nil {
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

type datasetListParams struct {
	OwnerID    *uint64
	CategoryID *uint64
	Public     *bool
	Search     string
	MinRating  *float64
	MaxRating  *float64
}

type datasetVisibility struct {
	ownerFilter  *uint64
	publicFilter *bool
	onlyPublic   bool
}

type datasetView struct {
	ds      *entities.Dataset
	summary services.RatingSummary
}

func parseDatasetListQuery(q url.Values) (datasetListParams, error) {
	ownerParam, err := parseUintPtr(q, "owner_id")
	if err != nil {
		return datasetListParams{}, err
	}
	categoryID, err := parseUintPtr(q, "category_id")
	if err != nil {
		return datasetListParams{}, err
	}
	publicFilter, err := parseBoolPtr(q, "is_public")
	if err != nil {
		return datasetListParams{}, err
	}
	minRating, err := parseRatingBound(q, "min_rating")
	if err != nil {
		return datasetListParams{}, err
	}
	maxRating, err := parseRatingBound(q, "max_rating")
	if err != nil {
		return datasetListParams{}, err
	}
	if minRating != nil && maxRating != nil && *minRating > *maxRating {
		return datasetListParams{}, &dto.BadRequestError{Message: "min_rating cannot be greater than max_rating"}
	}

	return datasetListParams{
		OwnerID:    ownerParam,
		CategoryID: categoryID,
		Public:     publicFilter,
		Search:     strings.ToLower(strings.TrimSpace(q.Get("search"))),
		MinRating:  minRating,
		MaxRating:  maxRating,
	}, nil
}

func resolveDatasetVisibility(actor *entities.User, ownerParam *uint64, publicFilter *bool) (datasetVisibility, error) {
	if actor == nil {
		if publicFilter != nil && !*publicFilter {
			publicFilter = nil
		}
		return datasetVisibility{
			ownerFilter:  nil,
			publicFilter: publicFilter,
			onlyPublic:   true,
		}, nil
	}

	ownerFilter, err := resolveOwnerFilter(actor, ownerParam)
	if err != nil {
		return datasetVisibility{}, err
	}

	requestedPrivate := publicFilter != nil && !*publicFilter
	if requestedPrivate && !isAdmin(actor) {
		if ownerFilter == nil {
			id := actor.ID
			ownerFilter = &id
		} else if *ownerFilter != actor.ID {
			return datasetVisibility{}, services.ErrRequestForbidden
		}
	}

	return datasetVisibility{
		ownerFilter:  ownerFilter,
		publicFilter: publicFilter,
		onlyPublic:   false,
	}, nil
}

func resolveOwnerFilter(actor *entities.User, ownerParam *uint64) (*uint64, error) {
	if ownerParam == nil {
		return nil, nil
	}
	if isAdmin(actor) || *ownerParam == actor.ID {
		return ownerParam, nil
	}
	return nil, services.ErrRequestForbidden
}

func (h *Handler) collectDatasetViews(ctx context.Context, actor *entities.User, datasets []*entities.Dataset, params datasetListParams, publicFilter *bool) ([]datasetView, error) {
	views := make([]datasetView, 0, len(datasets))
	for _, ds := range datasets {
		if !datasetPassesFilters(ds, params, publicFilter) {
			continue
		}
		view, err := h.buildDatasetView(ctx, actor, ds, params)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) || errors.Is(err, services.ErrRequestForbidden) {
				continue
			}
			return nil, err
		}
		if view != nil {
			views = append(views, *view)
		}
	}
	return views, nil
}

func datasetPassesFilters(ds *entities.Dataset, params datasetListParams, publicFilter *bool) bool {
	if params.CategoryID != nil && ds.CategoryID != *params.CategoryID {
		return false
	}
	if publicFilter != nil && ds.IsPublic != *publicFilter {
		return false
	}
	if params.Search != "" && !matchesSearch(ds, params.Search) {
		return false
	}
	return true
}

func (h *Handler) buildDatasetView(ctx context.Context, actor *entities.User, ds *entities.Dataset, params datasetListParams) (*datasetView, error) {
	if err := ensureDatasetReadable(actor, ds.OwnerID, ds.IsPublic); err != nil {
		return nil, err
	}
	summary, err := h.reviews.GetRatingSummary(ctx, ds.ID)
	if err != nil {
		return nil, err
	}
	if !matchesRating(summary, params.MinRating, params.MaxRating) {
		return nil, nil
	}
	return &datasetView{ds: ds, summary: summary}, nil
}

func paginateDatasets(views []datasetView, limit, offset int, h *Handler) []DatasetResponse {
	total := len(views)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)

	items := make([]DatasetResponse, 0, end-start)
	for _, view := range views[start:end] {
		items = append(items, h.toDatasetResponse(view.ds, view.summary, nil))
	}
	return items
}

func matchesRating(summary services.RatingSummary, minRating, maxRating *float64) bool {
	if minRating != nil && summary.Average < *minRating {
		return false
	}
	if maxRating != nil && summary.Average > *maxRating {
		return false
	}
	return true
}

func matchesSearch(ds *entities.Dataset, search string) bool {
	lowerName := strings.ToLower(ds.Name)
	lowerDesc := strings.ToLower(ds.Description)
	return strings.Contains(lowerName, search) || strings.Contains(lowerDesc, search)
}

func parseCreateDatasetRequest(r *http.Request, actorID uint64) (services.CreateDatasetCmd, multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return services.CreateDatasetCmd{}, nil, nil, &dto.BadRequestError{Message: "invalid multipart form"}
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return services.CreateDatasetCmd{}, nil, nil, &dto.BadRequestError{Message: "name is required"}
	}
	categoryRaw := firstNonEmpty(r.FormValue("category_id"), r.FormValue("category"))
	catID, err := strconv.ParseUint(categoryRaw, 10, 64)
	if err != nil || catID == 0 {
		return services.CreateDatasetCmd{}, nil, nil, &dto.BadRequestError{Message: "category_id must be a positive integer"}
	}

	publicRaw := firstNonEmpty(r.FormValue("is_public"), r.FormValue("public"))
	isPublic := false
	if publicRaw != "" {
		val, parseErr := strconv.ParseBool(publicRaw)
		if parseErr != nil {
			return services.CreateDatasetCmd{}, nil, nil, &dto.BadRequestError{Message: "is_public must be boolean"}
		}
		isPublic = val
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return services.CreateDatasetCmd{}, nil, nil, &dto.BadRequestError{Message: "file is required"}
	}

	metaFormat := firstNonEmpty(r.FormValue("meta_format"), r.FormValue("metaFormat"))
	metaTags := collectMultipartTags(r.MultipartForm, "meta_tags", "metaTags")
	if err := mergeMetadataJSON(strings.TrimSpace(r.FormValue("metadata")), &metaFormat, &metaTags); err != nil {
		return services.CreateDatasetCmd{}, nil, nil, &dto.BadRequestError{Message: "metadata must be valid JSON"}
	}

	cmd := services.CreateDatasetCmd{
		ActorID:     actorID,
		Name:        name,
		Description: r.FormValue("description"),
		CategoryID:  catID,
		IsPublic:    isPublic,
		FileName:    header.Filename,
		MetaFormat:  metaFormat,
		MetaTags:    metaTags,
		MetaSize:    uint64(header.Size),
	}
	return cmd, file, header, nil
}

func parseDatasetVersionRequest(r *http.Request, actorID, datasetID uint64) (services.AddVersionCmd, multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		return services.AddVersionCmd{}, nil, nil, &dto.BadRequestError{Message: "invalid multipart form"}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return services.AddVersionCmd{}, nil, nil, &dto.BadRequestError{Message: "file is required"}
	}

	cmd := services.AddVersionCmd{
		ActorID:    actorID,
		DatasetID:  datasetID,
		ChangeLog:  r.FormValue("change_log"),
		FileName:   header.Filename,
		MetaFormat: firstNonEmpty(r.FormValue("meta_format"), r.FormValue("metaFormat")),
		MetaTags:   collectMultipartTags(r.MultipartForm, "meta_tags", "metaTags"),
		MetaSize:   uint64(header.Size),
	}
	if err := mergeMetadataJSON(strings.TrimSpace(r.FormValue("metadata")), &cmd.MetaFormat, &cmd.MetaTags); err != nil {
		return services.AddVersionCmd{}, nil, nil, &dto.BadRequestError{Message: "metadata must be valid JSON"}
	}

	return cmd, file, header, nil
}

type updateDatasetRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	CategoryID  *uint64 `json:"category_id"`
	IsPublic    *bool   `json:"is_public"`
}

func decodeUpdateDatasetRequest(r *http.Request) (updateDatasetRequest, error) {
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		CategoryID  *uint64 `json:"category_id"`
		IsPublic    *bool   `json:"is_public"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return updateDatasetRequest{}, err
	}
	if req.Name == nil && req.Description == nil && req.CategoryID == nil && req.IsPublic == nil {
		return updateDatasetRequest{}, &dto.BadRequestError{Message: "nothing to update"}
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return updateDatasetRequest{}, &dto.BadRequestError{Message: "name cannot be empty"}
		}
	}
	if req.CategoryID != nil {
		if *req.CategoryID == 0 {
			return updateDatasetRequest{}, &dto.BadRequestError{Message: "category_id must be positive"}
		}
	}
	return updateDatasetRequest{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		IsPublic:    req.IsPublic,
	}, nil
}

func buildUpdateDatasetCmd(id uint64, current *entities.Dataset, updates updateDatasetRequest) (services.UpdateDatasetCmd, error) {
	cmd := services.UpdateDatasetCmd{
		ID:          id,
		Name:        current.Name,
		Description: current.Description,
		CategoryID:  current.CategoryID,
		IsPublic:    current.IsPublic,
	}
	if updates.Name != nil {
		cmd.Name = *updates.Name
	}
	if updates.Description != nil {
		cmd.Description = *updates.Description
	}
	if updates.CategoryID != nil {
		cmd.CategoryID = *updates.CategoryID
	}
	if updates.IsPublic != nil {
		cmd.IsPublic = *updates.IsPublic
	}
	return cmd, nil
}

func (h *Handler) fullDatasetResponse(ctx context.Context, id uint64) (DatasetResponse, error) {
	ds, err := h.datasets.GetDataset(ctx, id)
	if err != nil {
		return DatasetResponse{}, err
	}
	summary, err := h.reviews.GetRatingSummary(ctx, id)
	if err != nil {
		return DatasetResponse{}, err
	}
	latest, err := h.latestVersion(ctx, id)
	if err != nil && !errors.Is(err, services.ErrVersionNotFound) {
		return DatasetResponse{}, err
	}
	var latestResp *DatasetVersionResponse
	if latest != nil {
		latestResp, err = h.buildVersionResponse(ctx, latest)
		if err != nil {
			return DatasetResponse{}, err
		}
	}
	return h.toDatasetResponse(ds, summary, latestResp), nil
}

func (h *Handler) loadReadableVersion(ctx context.Context, actor *entities.User, datasetID, versionID uint64) (*entities.DatasetVersion, error) {
	ds, err := h.datasets.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	if err := ensureDatasetReadable(actor, ds.OwnerID, ds.IsPublic); err != nil {
		return nil, err
	}

	version, err := h.datasets.GetVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version.DatasetID != datasetID {
		return nil, services.ErrVersionNotFound
	}
	return version, nil
}

func streamDatasetVersion(ctx context.Context, w http.ResponseWriter, url string, fallbackName string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("storage returned status %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		w.Header().Set("Content-Disposition", cd)
	} else if fallbackName != "" {
		w.Header().Set("Content-Disposition", `attachment; filename="`+fallbackName+`"`)
	}

	_, copyErr := io.Copy(w, resp.Body)
	if copyErr != nil {
		return nil
	}
	return nil
}

func (h *Handler) buildVersionResponse(ctx context.Context, ver *entities.DatasetVersion) (*DatasetVersionResponse, error) {
	if ver == nil {
		return nil, nil
	}
	md, err := h.datasets.GetVersionMetadata(ctx, ver.ID)
	if err != nil {
		return nil, err
	}
	resp := &DatasetVersionResponse{
		ID:        ver.ID,
		DatasetID: ver.DatasetID,
		Number:    ver.Number,
		ChangeLog: ver.ChangeLog,
		FileURL:   ver.Filepath,
		UploadAt:  ver.UploadDate,
		Metadata:  toDatasetMetadata(md),
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

func mergeMetadataJSON(raw string, format *string, tags *string) error {
	if raw == "" {
		return nil
	}
	var payload struct {
		Format string   `json:"format"`
		Tags   []string `json:"tags"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return err
	}
	if payload.Format != "" && format != nil {
		*format = payload.Format
	}
	if len(payload.Tags) > 0 && tags != nil {
		*tags = strings.Join(payload.Tags, ",")
	}
	return nil
}
