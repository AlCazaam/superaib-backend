package repo

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"op"`
	Value    interface{} `json:"value"`
}

type DocumentRepository interface {
	// --- 11 ADVANCED CRUD (The Powerhouse) ---
	Create(ctx context.Context, doc *models.Document) error                                                              // 1. Add
	GetByID(ctx context.Context, pID string, cID uuid.UUID, id string) (*models.Document, error)                         // 2. Get Single
	Set(ctx context.Context, doc *models.Document, merge bool) error                                                     // 3. Set (Overwrite/Merge)
	Update(ctx context.Context, pID string, cID uuid.UUID, id string, data map[string]interface{}, oldEtag string) error // 4. Update (Partial)
	Upsert(ctx context.Context, doc *models.Document) error                                                              // 5. Upsert
	Delete(ctx context.Context, pID string, cID uuid.UUID, id string) error                                              // 6. Delete
	Exists(ctx context.Context, pID string, cID uuid.UUID, id string) (bool, error)                                      // 7. Exists
	Count(ctx context.Context, pID string, cID uuid.UUID, filters []Filter) (int64, error)                               // 8. Count
	Increment(ctx context.Context, pID string, cID uuid.UUID, id, field string, amount float64) error                    // 9. Increment
	EnsureCollectionExists(ctx context.Context, pID, name string) (*models.Collection, error)                            // 10. Collection Create
	GetCollections(ctx context.Context, pID string) ([]models.Collection, error)                                         // 11. Collection List

	// --- 7 COMPREHENSIVE QUERY & FILTERING (The Magic) ---
	// 1. Where, 2. OrderBy, 3. Limit, 4. Offset, 5. Search, 6. Select, 7. Advanced Ops (In/Contains)
	QueryAdvanced(ctx context.Context, pID string, cID uuid.UUID, filters []Filter, selectFields []string, limit, offset int, orderBy string, search string) ([]models.Document, error)

	// Collection Management
	GetCollectionByName(ctx context.Context, projectID, name string) (*models.Collection, error)
	RenameCollection(ctx context.Context, projectID, collectionID, newName string) error
	DeleteCollection(ctx context.Context, projectID, collectionID string) error
}

type documentRepository struct{ db *gorm.DB }

func NewDocumentRepository(db *gorm.DB) DocumentRepository { return &documentRepository{db: db} }

// ... (Create, GetByID, Set, Upsert, Delete, Exists, Increment - Sideedii hore u deji) ...

func (r *documentRepository) Create(ctx context.Context, doc *models.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *documentRepository) GetByID(ctx context.Context, pID string, cID uuid.UUID, id string) (*models.Document, error) {
	var doc models.Document
	err := r.db.WithContext(ctx).Where("project_id = ? AND collection_id = ? AND id = ? AND is_deleted = false", pID, cID, id).First(&doc).Error
	return &doc, err
}

func (r *documentRepository) Set(ctx context.Context, doc *models.Document, merge bool) error {
	if !merge {
		return r.db.WithContext(ctx).Save(doc).Error
	}
	sql := `UPDATE documents SET data = data || ?, etag = ?, version = version + 1, updated_at = ? WHERE project_id = ? AND collection_id = ? AND id = ?`
	return r.db.WithContext(ctx).Exec(sql, doc.Data, uuid.New().String(), time.Now(), doc.ProjectID, doc.CollectionID, doc.ID).Error
}

func (r *documentRepository) Update(ctx context.Context, pID string, cID uuid.UUID, id string, data map[string]interface{}, oldEtag string) error {
	// 1. Dhis SQL-ka asaasiga ah (Merge JSONB xogta jirta iyo tan cusub)
	sql := `UPDATE documents 
	        SET data = data || ?, 
	            etag = ?, 
	            version = version + 1, 
	            updated_at = ? 
	        WHERE project_id = ? AND collection_id = ? AND id = ?`

	newEtag := uuid.New().String()
	now := time.Now()

	// 2. ✅ SMART LOGIC: Haddii Developer-ku uusan soo dirin ETag, ha weydiin.
	// Kaliya haddii uu si ula kac ah u soo diro (Security ahaan) ayaan hubineynaa.
	if oldEtag != "" && oldEtag != "null" {
		sql += fmt.Sprintf(" AND etag = '%s'", oldEtag)
	}

	result := r.db.WithContext(ctx).Exec(sql, data, newEtag, now, pID, cID, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("document_not_found")
	}
	return nil
}
func (r *documentRepository) Upsert(ctx context.Context, doc *models.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *documentRepository) Delete(ctx context.Context, pID string, cID uuid.UUID, id string) error {
	return r.db.WithContext(ctx).Model(&models.Document{}).Where("project_id = ? AND collection_id = ? AND id = ?", pID, cID, id).Update("is_deleted", true).Error
}

func (r *documentRepository) Exists(ctx context.Context, pID string, cID uuid.UUID, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Document{}).Where("project_id = ? AND collection_id = ? AND id = ? AND is_deleted = false", pID, cID, id).Count(&count).Error
	return count > 0, err
}

func (r *documentRepository) Count(ctx context.Context, pID string, cID uuid.UUID, filters []Filter) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Document{}).Where("project_id = ? AND collection_id = ? AND is_deleted = false", pID, cID)
	for _, f := range filters {
		q = applyFilter(q, f)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *documentRepository) Increment(ctx context.Context, pID string, cID uuid.UUID, id, field string, amount float64) error {
	sql := fmt.Sprintf(`UPDATE documents SET data = jsonb_set(data, '{%s}', (COALESCE(data->>'%s', '0')::numeric + ?)::text::jsonb) WHERE project_id = ? AND collection_id = ? AND id = ?`, field, field)
	return r.db.WithContext(ctx).Exec(sql, amount, pID, cID, id).Error
}

// 🚀 THE MAGIC: QueryAdvanced oo leh SELECT PROJECTION
// 🚀 THE MASTER QUERY: QueryAdvanced oo leh Full Logic (Select, Filter, Search, Order, Limit)
func (r *documentRepository) QueryAdvanced(ctx context.Context, pID string, cID uuid.UUID, filters []Filter, selectFields []string, limit, offset int, orderBy string, search string) ([]models.Document, error) {
	var docs []models.Document

	// 1. Bilow Query-ga asaasiga ah
	q := r.db.WithContext(ctx).Where("project_id = ? AND collection_id = ? AND is_deleted = false", pID, cID)

	// 2. ✅ SELECTION LOGIC: Kaliya soo saar xogta loo baahanyahay (Performance Booster)
	if len(selectFields) > 0 {
		// Waxaan dhisaynaa xariiq SQL ah oo dib u dhisaysa JSON-ka (Data field)
		// Metadata-da muhiimka ah (ID, CreatedAt, iwm) had iyo jeer waa inay soo baxaan
		projection := "id, project_id, collection_id, etag, version, created_at, updated_at, jsonb_build_object("

		for i, field := range selectFields {
			projection += fmt.Sprintf("'%s', data->'%s'", field, field)
			if i < len(selectFields)-1 {
				projection += ", "
			}
		}
		projection += ") as data"

		// Ku dar Select-ka Query-ga
		q = q.Select(projection)
	}

	// 3. ✅ GLOBAL TEXT SEARCH: Ka baar dhamaan JSON-ka dhexdiisa (Case Insensitive)
	if search != "" {
		q = q.Where("data::text ILIKE ?", "%"+search+"%")
	}

	// 4. ✅ DYNAMIC FILTERS: Codso dhamaan sifeeyayaasha (applyFilter helper)
	for _, f := range filters {
		q = applyFilter(q, f)
	}

	// 5. ✅ ORDERING: Habaynta xogta
	if orderBy == "" {
		orderBy = "created_at DESC"
	}

	// 6. ✅ EXECUTION: Ku dar Limit iyo Offset (Pagination)
	err := q.Order(orderBy).Limit(limit).Offset(offset).Find(&docs).Error

	return docs, err
}
func applyFilter(q *gorm.DB, f Filter) *gorm.DB {
	switch f.Operator {
	case "==", "=":
		return q.Where(fmt.Sprintf("data->>'%s' = ?", f.Field), f.Value)
	case "!=":
		return q.Where(fmt.Sprintf("data->>'%s' != ?", f.Field), f.Value)
	case ">":
		return q.Where(fmt.Sprintf("(data->>'%s')::numeric > ?", f.Field), f.Value)
	case "<":
		return q.Where(fmt.Sprintf("(data->>'%s')::numeric < ?", f.Field), f.Value)
	case ">=":
		return q.Where(fmt.Sprintf("(data->>'%s')::numeric >= ?", f.Field), f.Value)
	case "<=":
		return q.Where(fmt.Sprintf("(data->>'%s')::numeric <= ?", f.Field), f.Value)
	case "in":
		return q.Where(fmt.Sprintf("data->>'%s' IN ?", f.Field), f.Value) // Magic: In Operator
	case "contains":
		return q.Where(fmt.Sprintf("data->>'%s' ILIKE ?", f.Field), "%"+fmt.Sprint(f.Value)+"%")
	case "startsWith":
		return q.Where(fmt.Sprintf("data->>'%s' ILIKE ?", f.Field), fmt.Sprint(f.Value)+"%")
	default:
		return q.Where(fmt.Sprintf("data->>'%s' = ?", f.Field), f.Value)
	}
}

// ... (Collection operations sideedii deji) ...
func (r *documentRepository) EnsureCollectionExists(ctx context.Context, pID, name string) (*models.Collection, error) {
	coll, err := r.GetCollectionByName(ctx, pID, name)
	if err == nil {
		return coll, nil
	}
	newColl := &models.Collection{ID: uuid.New(), ProjectID: pID, Name: name}
	err = r.db.WithContext(ctx).Create(newColl).Error
	return newColl, err
}
func (r *documentRepository) GetCollections(ctx context.Context, projectID string) ([]models.Collection, error) {
	var colls []models.Collection
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&colls).Error
	return colls, err
}
func (r *documentRepository) GetCollectionByName(ctx context.Context, projectID, name string) (*models.Collection, error) {
	var coll models.Collection
	err := r.db.WithContext(ctx).Where("project_id = ? AND name = ?", projectID, name).First(&coll).Error
	return &coll, err
}
func (r *documentRepository) RenameCollection(ctx context.Context, projectID, collectionID, newName string) error {
	return r.db.WithContext(ctx).Model(&models.Collection{}).Where("project_id = ? AND id = ?", projectID, collectionID).Update("name", newName).Error
}
func (r *documentRepository) DeleteCollection(ctx context.Context, projectID, collectionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx.Model(&models.Document{}).Where("project_id = ? AND collection_id = ?", projectID, collectionID).Update("is_deleted", true)
		return tx.Where("project_id = ? AND id = ?", projectID, collectionID).Delete(&models.Collection{}).Error
	})
}
