package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"superaib/internal/models"
	"superaib/internal/storage/repo"

	"github.com/google/uuid"
)

type StorageService interface {
	CreateFile(ctx context.Context, file *models.StorageFile) (*models.StorageFile, error)
	GetFilesByProject(ctx context.Context, projectID string, page, pageSize int) ([]models.StorageFile, int64, error)
	GetFileByID(ctx context.Context, fileID string) (*models.StorageFile, error)
	DeleteFile(ctx context.Context, fileID string) error
	UploadToCloud(ctx context.Context, projectID string, file []byte, fileName string, fileType string) (*models.StorageFile, error)
}

type storageService struct {
	repo         repo.StorageRepository
	featureRepo  repo.ProjectFeatureRepository
	tracker      *AnalyticsTracker
	usageService ProjectUsageService // ✅ KU DAR: Si aan u xino limits-ka
}

// ✅ Constructor-ka hadda wuxuu aqbalayaa 4 dependencies
func NewStorageService(r repo.StorageRepository, fr repo.ProjectFeatureRepository, tracker *AnalyticsTracker, usage ProjectUsageService) StorageService {
	return &storageService{
		repo:         r,
		featureRepo:  fr,
		tracker:      tracker,
		usageService: usage,
	}
}

func (s *storageService) CreateFile(ctx context.Context, file *models.StorageFile) (*models.StorageFile, error) {
	// 1. Hubi haddii Storage uu "Enabled" u yahay
	feature, err := s.featureRepo.GetFeatureByProjectIDAndType(ctx, file.ProjectID, models.FeatureTypeStorage)
	if err != nil || !feature.Enabled {
		return nil, errors.New("storage feature is disabled for this project")
	}

	// 2. Keydi xogta file-ka ee Database-ka
	if err := s.repo.Create(ctx, file); err != nil {
		return nil, err
	}

	// ✅ 3. TRACK ANALYTICS (Monthly Trends/Charts)
	s.tracker.TrackEvent(ctx, file.ProjectID, models.AnalyticsTypeStorageUsage, "files_uploaded", 1)
	s.tracker.TrackEvent(ctx, file.ProjectID, models.AnalyticsTypeStorageUsage, "total_storage_mb", file.SizeMB)

	// ✅ 4. UPDATE PROJECT USAGE (Si loogu xakameeyo Limits-ka Plan-ka)
	// Field-ka halkan waa 'storage_used_mb' sidii uu ahaa models.ProjectUsage
	_ = s.usageService.UpdateUsage(ctx, file.ProjectID, "storage_used_mb", file.SizeMB)

	return file, nil
}

func (s *storageService) GetFilesByProject(ctx context.Context, projectID string, page, pageSize int) ([]models.StorageFile, int64, error) {
	return s.repo.GetByProject(ctx, projectID, page, pageSize)
}

func (s *storageService) GetFileByID(ctx context.Context, fileID string) (*models.StorageFile, error) {
	id, err := uuid.Parse(fileID)
	if err != nil {
		return nil, errors.New("invalid file id")
	}
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// ✅ TRACK: Analytics kaliya (Akhris maku jiro Usage limits)
	s.tracker.TrackEvent(ctx, file.ProjectID, models.AnalyticsTypeStorageUsage, "files_read", 1)
	return file, nil
}

func (s *storageService) DeleteFile(ctx context.Context, fileID string) error {
	id, err := uuid.Parse(fileID)
	if err != nil {
		return errors.New("invalid file id")
	}

	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// ✅ 1. TRACK ANALYTICS (Marka la tirtiro)
	s.tracker.TrackEvent(ctx, file.ProjectID, models.AnalyticsTypeStorageUsage, "files_deleted", 1)
	s.tracker.TrackEvent(ctx, file.ProjectID, models.AnalyticsTypeStorageUsage, "total_storage_mb", -file.SizeMB)

	// ✅ 2. UPDATE PROJECT USAGE (Ka dhim MB-yada mashruuca hadda u xareysan)
	// Waxaan u dhiibaynaa '-' si uu Database-ka uga jaro (Decrement)
	_ = s.usageService.UpdateUsage(ctx, file.ProjectID, "storage_used_mb", -file.SizeMB)

	return nil
}
func (s *storageService) UploadToCloud(ctx context.Context, projectID string, fileData []byte, fileName string, fileType string) (*models.StorageFile, error) {
	// 1. Hel Feature Config-ga (Kii Developer-ku gashaday - Image 1)
	feature, err := s.featureRepo.GetFeatureByProjectIDAndType(ctx, projectID, models.FeatureTypeStorage)
	if err != nil || !feature.Enabled {
		return nil, errors.New("storage feature is disabled for this project")
	}

	// 2. Kasoo saar Config-ga (Cloud Name iyo Upload Preset)
	var config map[string]string
	if err := json.Unmarshal(feature.Config, &config); err != nil {
		return nil, errors.New("invalid storage configuration")
	}

	cloudName := config["cloud_name"]
	uploadPreset := config["upload_preset"]

	if cloudName == "" || uploadPreset == "" {
		return nil, errors.New("cloudinary config (cloud_name or upload_preset) is missing")
	}

	// 🚀 3. UPLOAD TO CLOUDINARY (REAL API CALL)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Ku dar File-ka
	part, _ := writer.CreateFormFile("file", fileName)
	part.Write(fileData)

	// Ku dar Upload Preset (Khasab ah Cloudinary)
	writer.WriteField("upload_preset", uploadPreset)
	writer.Close()

	cloudinaryURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	req, _ := http.NewRequest("POST", cloudinaryURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudinary request failed: %v", err)
	}
	defer resp.Body.Close()

	// 4. Akhri Jawaabta Cloudinary
	var result map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &result)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloudinary upload failed: %v", result["error"])
	}

	// Hel URL-ka rasmiga ah (Secure URL)
	finalURL := result["secure_url"].(string)

	// 5. Kaydi Metadata-ga pgAdmin dhexdiisa
	newFile := &models.StorageFile{
		ProjectID: projectID,
		FileName:  fileName,
		FileType:  fileType,
		SizeMB:    float64(len(fileData)) / (1024 * 1024),
		URL:       finalURL, // 👈 Hadda waa URL dhab ah!
	}

	if err := s.repo.Create(ctx, newFile); err != nil {
		return nil, err
	}

	// Analytics
	s.tracker.TrackEvent(ctx, projectID, models.AnalyticsTypeStorageUsage, "files_uploaded", 1)
	s.usageService.UpdateUsage(ctx, projectID, "storage_used_mb", newFile.SizeMB)

	return newFile, nil
}
