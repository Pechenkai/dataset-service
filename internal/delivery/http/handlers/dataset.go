package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

// @Summary      List datasets
// @Description  Returns datasets, optionally filtering by public or owner
// @Tags         datasets
// @Produce      json
// @Param        public  query     bool   false  "Only public"
// @Param        user    query     int    false  "Owner user ID"
// @Success      200     {object}  dto.DatasetsResponse
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /datasets [get]
func ListDatasets(svc services.DatasetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		publicOnly := false
		if p := q.Get("public"); p != "" {
			publicOnly, _ = strconv.ParseBool(p)
		}
		var ownerID *uint64
		if u := q.Get("user"); u != "" {
			id, err := strconv.ParseUint(u, 10, 64)
			if err == nil {
				ownerID = &id
			}
		}

		list, err := svc.ListDatasets(r.Context(), publicOnly, ownerID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		dto.WriteJSON(w, http.StatusOK, dto.FromDatasetList(list))
	}
}

// @Summary      Get dataset by ID
// @Description  Returns a single dataset
// @Tags         datasets
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  dto.DatasetResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /datasets/{id} [get]
func GetDataset(svc services.DatasetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		d, err := svc.GetDataset(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		dto.WriteJSON(w, http.StatusOK, dto.FromDataset(d))
	}
}

// @Summary      Create a new dataset
// @Description  Uploads dataset and initial version
// @Tags         datasets
// @Accept       multipart/form-data
// @Produce      json
// @Param        name        formData  string  true   "Dataset name"
// @Param        description formData  string  false  "Dataset description"
// @Param        category    formData  int     true   "Category ID"
// @Param        public      formData  bool    false  "Is public"
// @Param        file        formData  file    true   "Data file to upload"
// @Param        metaFormat  formData  string  false  "Metadata format"
// @Param        metaTags    formData  string  false  "Metadata tags"
// @Param        metaSize    formData  int     false  "Metadata size"
// @Success      201         {object}  dto.DatasetResponse
// @Failure      400         {object}  dto.ErrorResponse
// @Failure      500         {object}  dto.ErrorResponse
// @Router       /datasets [post]
func CreateDataset(svc services.DatasetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			dto.WriteError(w, err)
			return
		}
		name := r.FormValue("name")
		desc := r.FormValue("description")
		catID, _ := strconv.ParseUint(r.FormValue("category"), 10, 64)
		isPub, _ := strconv.ParseBool(r.FormValue("public"))
		metaFmt := r.FormValue("metaFormat")
		tags := r.FormValue("metaTags")
		sz, _ := strconv.ParseUint(r.FormValue("metaSize"), 10, 64)

		// file
		file, header, err := r.FormFile("file")
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		defer file.Close()
		size := header.Size

		cmd := services.CreateDatasetCmd{
			Name:        name,
			Description: desc,
			CategoryID:  catID,
			IsPublic:    isPub,
			FileName:    header.Filename,
			MetaFormat:  metaFmt,
			MetaTags:    tags,
			MetaSize:    sz,
		}

		id, err := svc.CreateDataset(r.Context(), cmd, file.(io.Reader), size)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		d, err := svc.GetDataset(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		dto.WriteJSON(w, http.StatusCreated, dto.FromDataset(d))
	}
}

// @Summary      Add dataset version
// @Description  Uploads a new version file for a dataset
// @Tags         datasets
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path      int     true   "Dataset ID"
// @Param        file        formData  file    true   "Data file to upload"
// @Param        changeLog   formData  string  false  "Change log"
// @Param        metaFormat  formData  string  false  "Metadata format"
// @Param        metaTags    formData  string  false  "Metadata tags"
// @Param        metaSize    formData  int     false  "Metadata size"
// @Success      201         {object}  dto.VersionResponse
// @Failure      400         {object}  dto.ErrorResponse
// @Failure      500         {object}  dto.ErrorResponse
// @Router       /datasets/{id}/versions [post]
func AddVersion(svc services.DatasetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		dsID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			dto.WriteError(w, err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		defer file.Close()

		size := header.Size
		chLog := r.FormValue("changeLog")
		metaFmt := r.FormValue("metaFormat")
		tags := r.FormValue("metaTags")
		sz, _ := strconv.ParseUint(r.FormValue("metaSize"), 10, 64)

		cmd := services.AddVersionCmd{
			DatasetID:  dsID,
			ChangeLog:  chLog,
			FileName:   header.Filename,
			MetaFormat: metaFmt,
			MetaTags:   tags,
			MetaSize:   sz,
		}

		vid, err := svc.AddDatasetVersion(r.Context(), cmd, file.(io.Reader), size)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		ver, err := svc.GetVersion(r.Context(), vid)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusCreated, dto.FromVersion(ver))
	}
}

// @Summary      List dataset versions
// @Description  Returns all versions for a given dataset
// @Tags         datasets
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  dto.VersionsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /datasets/{id}/versions [get]
func ListVersions(svc services.DatasetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid dataset ID"})
			return
		}

		versions, err := svc.ListVersions(r.Context(), datasetID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.FromVersionList(versions)

		dto.WriteJSON(w, http.StatusOK, resp)
	}
}
