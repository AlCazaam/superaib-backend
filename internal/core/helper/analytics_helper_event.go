package helper

import (
	"context"
	"encoding/json"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"gorm.io/gorm"
)

// AnalyticsHelper wuxuu caawiyaa in xogta analytics-ka la raad-raaco iyadoo la dhex jiro Transaction.
type AnalyticsHelper struct {
	repo repo.AnalyticsRepository
}

func NewAnalyticsHelper(repo repo.AnalyticsRepository) *AnalyticsHelper {
	return &AnalyticsHelper{repo: repo}
}

// TrackEvent: Kani wuxuu si toos ah database-ka u cusbooneysiinayaa (Sync tracking).
func (h *AnalyticsHelper) TrackEvent(ctx context.Context, tx *gorm.DB, projectID string, aType models.AnalyticsType, data map[string]interface{}) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Fetch existing records for the month
	records, err := h.repo.GetByProjectForPeriod(ctx, projectID, periodStart)
	if err != nil {
		return err
	}

	var rec *models.Analytics
	for i := range records {
		if records[i].Type == aType {
			rec = &records[i]
			break
		}
	}

	// Haddii record-ku uusan jirin, abuuro
	if rec == nil {
		metricsJSON, _ := json.Marshal(data)
		newRec := &models.Analytics{
			ProjectID:   projectID,
			Type:        aType,
			PeriodStart: periodStart,
			Metrics:     metricsJSON,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		return h.repo.Create(ctx, tx, newRec)
	}

	// Merge metrics (Isku dar xogta cusub iyo tan hore)
	var existing map[string]interface{}
	_ = json.Unmarshal(rec.Metrics, &existing)

	for k, v := range data {
		if val, ok := existing[k].(float64); ok {
			if newVal, ok := v.(float64); ok {
				existing[k] = val + newVal
			} else if newVal, ok := v.(int); ok {
				existing[k] = val + float64(newVal)
			}
		} else {
			existing[k] = v
		}
	}

	return nil
	//h.repo.UpdateMetrics(ctx, projectID, aType, periodStart, existing)
}
