package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type StorageHandler struct {
	service services.StorageService
}

func NewStorageHandler(s services.StorageService) *StorageHandler {
	return &StorageHandler{service: s}
}

// CreateFile handles POST /projects/{project_id}/storage/files
func (h *StorageHandler) CreateFile(w http.ResponseWriter, r *http.Request) {
	// Ka soo qaad project_id URL-ka (Waxaa soo saaray Middleware-ka)
	projectID := mux.Vars(r)["project_id"]

	var file models.StorageFile
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Hubi in project_id uu sax yahay
	file.ProjectID = projectID

	createdFile, err := h.service.CreateFile(r.Context(), &file)
	if err != nil {
		// Haddii uu feature-ku xiran yahay, halkan ayuu 403 ku soo celinayaa
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "File record created", createdFile)
}

// GetFiles handles GET /projects/{project_id}/storage/files
func (h *StorageHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	} // Good default for a grid view

	files, total, err := h.service.GetFilesByProject(r.Context(), projectID, page, pageSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get files", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Files retrieved", map[string]interface{}{
		"total": total, "page": page, "pageSize": pageSize, "files": files,
	})
}

// DeleteFile handles DELETE /storage/files/{id}
func (h *StorageHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]

	if err := h.service.DeleteFile(r.Context(), fileID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete file", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "File deleted successfully", nil)
}

// 🚀 UPLOAD FILE: Kani waa kan SDK-ga u jawaabaya
func (h *StorageHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]

	// 1. Parse Multipart Form (Sawirka dhabta ah)
	err := r.ParseMultipartForm(10 << 20) // Max 10MB
	if err != nil {
		response.Error(w, 400, "File too large or invalid")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, 400, "No file provided in 'file' field")
		return
	}
	defer file.Close()

	// 2. Akhri Bytes-ka sawirka
	fileBytes, _ := io.ReadAll(file)
	fileName := header.Filename
	fileType := header.Header.Get("Content-Type")

	// 3. U dir Service-ka si uu Cloudinary ugu rido
	createdFile, err := h.service.UploadToCloud(r.Context(), projectID, fileBytes, fileName, fileType)
	if err != nil {
		response.Error(w, 500, err.Error())
		return
	}

	fmt.Printf("✅ [Storage] File %s uploaded and saved to pgAdmin\n", fileName)
	response.JSON(w, 201, "Uploaded successfully", createdFile)
}
