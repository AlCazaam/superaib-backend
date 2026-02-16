package services

import (
	"context"
	"encoding/json"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AdvancedQueryRequest: 7 Query Features halkan ayay ku jiraan (Select, Limit, Offset, Order, Search)
type AdvancedQueryRequest struct {
	Filters      []repo.Filter `json:"filters"`
	SelectFields []string      `json:"select"`
	Limit        int           `json:"limit"`
	Offset       int           `json:"offset"`
	OrderBy      string        `json:"order_by"`
	Search       string        `json:"search"`
}

type DocumentService interface {
	// --- 11 CORE OPERATIONS ---
	Create(ctx context.Context, projectID, collectionName string, data map[string]interface{}) (*models.Document, error)
	Get(ctx context.Context, projectID, collectionName, id string) (*models.Document, error)
	Set(ctx context.Context, projectID, collectionName, id string, data map[string]interface{}, merge bool) (*models.Document, error)
	Update(ctx context.Context, projectID, collectionName, id string, data map[string]interface{}, etag string) (*models.Document, error)
	Upsert(ctx context.Context, projectID, collectionName, id string, data map[string]interface{}) (*models.Document, error)
	Delete(ctx context.Context, projectID, collectionName, id string) error
	Exists(ctx context.Context, projectID, collectionName, id string) (bool, error)
	Count(ctx context.Context, projectID, collectionName string, filters []repo.Filter) (int64, error)
	Increment(ctx context.Context, projectID, collectionName, id, field string, amount float64) error

	// --- QUERY INTERFACES ---
	Search(ctx context.Context, projectID, collectionName string, filters []repo.Filter, limit, offset int) ([]models.Document, error)
	AdvancedSearch(ctx context.Context, projectID, collectionName string, req AdvancedQueryRequest) ([]models.Document, error)

	// --- COLLECTION MANAGEMENT ---
	ListCollections(ctx context.Context, projectID string) ([]models.Collection, error)
	CreateColl(ctx context.Context, projectID, name string) (*models.Collection, error)
	RenameColl(ctx context.Context, projectID, collectionID, newName string) error
	DeleteColl(ctx context.Context, projectID, collectionID string) error
}

type documentService struct {
	repo         repo.DocumentRepository
	tracker      *AnalyticsTracker
	usageService ProjectUsageService
}

func NewDocumentService(r repo.DocumentRepository, tracker *AnalyticsTracker, usage ProjectUsageService) DocumentService {
	return &documentService{repo: r, tracker: tracker, usageService: usage}
}

func mapToJSON(m map[string]interface{}) datatypes.JSON {
	b, _ := json.Marshal(m)
	return datatypes.JSON(b)
}

// ... Implementation-ka waa midka saxda ah ee dhamaantood wacaya repo-ga ...

func (s *documentService) Create(ctx context.Context, pID, collName string, data map[string]interface{}) (*models.Document, error) {
	coll, err := s.repo.EnsureCollectionExists(ctx, pID, collName)
	if err != nil {
		return nil, err
	}
	doc := &models.Document{
		ID: uuid.New(), ProjectID: pID, CollectionID: coll.ID, Data: mapToJSON(data),
		Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, err
	}
	s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_writes", 1)
	_ = s.usageService.UpdateUsage(ctx, pID, "documents_count", 1)
	return doc, nil
}

func (s *documentService) Get(ctx context.Context, pID, collName, id string) (*models.Document, error) {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return nil, err
	}
	doc, err := s.repo.GetByID(ctx, pID, coll.ID, id)
	if err == nil {
		s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_reads", 1)
	}
	return doc, err
}

func (s *documentService) Set(ctx context.Context, pID, collName, id string, data map[string]interface{}, merge bool) (*models.Document, error) {
	coll, _ := s.repo.EnsureCollectionExists(ctx, pID, collName)
	parsedID, _ := uuid.Parse(id)
	doc := &models.Document{ID: parsedID, ProjectID: pID, CollectionID: coll.ID, Data: mapToJSON(data)}
	if err := s.repo.Set(ctx, doc, merge); err != nil {
		return nil, err
	}
	s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_writes", 1)
	return doc, nil
}

func (s *documentService) Update(ctx context.Context, pID, collName, id string, data map[string]interface{}, etag string) (*models.Document, error) {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, pID, coll.ID, id, data, etag); err != nil {
		return nil, err
	}
	s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_writes", 1)
	return s.repo.GetByID(ctx, pID, coll.ID, id)
}

func (s *documentService) Upsert(ctx context.Context, pID, collName, id string, data map[string]interface{}) (*models.Document, error) {
	coll, _ := s.repo.EnsureCollectionExists(ctx, pID, collName)
	parsedID, _ := uuid.Parse(id)
	doc := &models.Document{ID: parsedID, ProjectID: pID, CollectionID: coll.ID, Data: mapToJSON(data), UpdatedAt: time.Now()}
	if err := s.repo.Upsert(ctx, doc); err != nil {
		return nil, err
	}
	s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_writes", 1)
	return doc, nil
}

func (s *documentService) Delete(ctx context.Context, pID, collName, id string) error {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, pID, coll.ID, id); err != nil {
		return err
	}
	s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_deletes", 1)
	_ = s.usageService.UpdateUsage(ctx, pID, "documents_count", -1)
	return nil
}

func (s *documentService) Exists(ctx context.Context, pID, collName, id string) (bool, error) {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return false, err
	}
	return s.repo.Exists(ctx, pID, coll.ID, id)
}

func (s *documentService) Count(ctx context.Context, pID, collName string, filters []repo.Filter) (int64, error) {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return 0, err
	}
	return s.repo.Count(ctx, pID, coll.ID, filters)
}

func (s *documentService) Increment(ctx context.Context, pID, collName, id, field string, amount float64) error {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return err
	}
	err = s.repo.Increment(ctx, pID, coll.ID, id, field, amount)
	if err == nil {
		s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_writes", 1)
	}
	return err
}

func (s *documentService) Search(ctx context.Context, pID, collName string, filters []repo.Filter, limit, offset int) ([]models.Document, error) {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return nil, err
	}
	docs, err := s.repo.QueryAdvanced(ctx, pID, coll.ID, filters, nil, limit, offset, "", "")
	if err == nil {
		s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_reads", float64(len(docs)))
	}
	return docs, err
}

func (s *documentService) AdvancedSearch(ctx context.Context, pID, collName string, req AdvancedQueryRequest) ([]models.Document, error) {
	coll, err := s.repo.GetCollectionByName(ctx, pID, collName)
	if err != nil {
		return nil, err
	}
	docs, err := s.repo.QueryAdvanced(ctx, pID, coll.ID, req.Filters, req.SelectFields, req.Limit, req.Offset, req.OrderBy, req.Search)
	if err == nil {
		s.tracker.TrackEvent(ctx, pID, models.AnalyticsTypeDatabaseUsage, "doc_reads", float64(len(docs)))
	}
	return docs, err
}

// --- Collection Management ---
func (s *documentService) ListCollections(ctx context.Context, pID string) ([]models.Collection, error) {
	return s.repo.GetCollections(ctx, pID)
}
func (s *documentService) CreateColl(ctx context.Context, pID, name string) (*models.Collection, error) {
	return s.repo.EnsureCollectionExists(ctx, pID, name)
}
func (s *documentService) RenameColl(ctx context.Context, pID, cID, newName string) error {
	return s.repo.RenameCollection(ctx, pID, cID, newName)
}
func (s *documentService) DeleteColl(ctx context.Context, pID, cID string) error {
	return s.repo.DeleteCollection(ctx, pID, cID)
}
