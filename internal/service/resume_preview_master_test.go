package service

import (
	"testing"

	"cut-the-bs/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestInferPreviewMasterSummaryID_LabelMatch(t *testing.T) {
	summaries := []domain.ProfessionalSummary{
		{ID: 1, Label: "General", BodyText: "General summary"},
		{ID: 2, Label: "Master Summary", BodyText: "Master paragraph"},
		{ID: 3, Label: "Project", BodyText: "Bullet summary"},
	}

	id := inferPreviewMasterSummaryID(summaries)
	if assert.NotNil(t, id) {
		assert.Equal(t, int64(2), *id)
	}
}

func TestInferPreviewMasterSummaryID_FallbackFirstNonEmpty(t *testing.T) {
	summaries := []domain.ProfessionalSummary{
		{ID: 10, Label: "General", BodyText: "First non-empty"},
		{ID: 11, Label: "Another", BodyText: "Second summary"},
	}

	id := inferPreviewMasterSummaryID(summaries)
	if assert.NotNil(t, id) {
		assert.Equal(t, int64(10), *id)
	}
}

func TestInferPreviewMasterSummaryID_PrefersNonEmptyMaster(t *testing.T) {
	summaries := []domain.ProfessionalSummary{
		{ID: 21, Label: "Master", BodyText: ""},
		{ID: 22, Label: "Master Summary", BodyText: "Preferred master"},
		{ID: 23, Label: "General", BodyText: "General"},
	}

	id := inferPreviewMasterSummaryID(summaries)
	if assert.NotNil(t, id) {
		assert.Equal(t, int64(22), *id)
	}
}

func TestInferPreviewMasterSummaryID_EmptyInput(t *testing.T) {
	assert.Nil(t, inferPreviewMasterSummaryID(nil))
}
