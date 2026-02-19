package service

import (
	"context"
	"fmt"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAcademicStore implements AcademicStore for testing.
type mockAcademicStore struct {
	credentials []domain.AcademicCredential
	credential  domain.AcademicCredential
	certs       []domain.Certification
	cert        domain.Certification
	err         error

	// call tracking
	createCredCalls  []domain.AcademicInput
	updateCredCalls  []updateCredCall
	deleteCredCalls  []int64
	reorderCredCalls [][]int64
	createCertCalls  []domain.CertificationInput
	updateCertCalls  []updateCertCall
	deleteCertCalls  []int64
	reorderCertCalls [][]int64
}

type updateCredCall struct {
	ID    int64
	Input domain.AcademicInput
}

type updateCertCall struct {
	ID    int64
	Input domain.CertificationInput
}

func (m *mockAcademicStore) ListAcademicCredentials(_ context.Context) ([]domain.AcademicCredential, error) {
	return m.credentials, m.err
}

func (m *mockAcademicStore) CreateAcademicCredential(_ context.Context, input domain.AcademicInput) (domain.AcademicCredential, error) {
	m.createCredCalls = append(m.createCredCalls, input)
	if m.err != nil {
		return domain.AcademicCredential{}, m.err
	}
	return m.credential, nil
}

func (m *mockAcademicStore) UpdateAcademicCredential(_ context.Context, id int64, input domain.AcademicInput) (domain.AcademicCredential, error) {
	m.updateCredCalls = append(m.updateCredCalls, updateCredCall{ID: id, Input: input})
	if m.err != nil {
		return domain.AcademicCredential{}, m.err
	}
	return m.credential, nil
}

func (m *mockAcademicStore) DeleteAcademicCredential(_ context.Context, id int64) error {
	m.deleteCredCalls = append(m.deleteCredCalls, id)
	return m.err
}

func (m *mockAcademicStore) ReorderAcademicCredentials(_ context.Context, orderedIDs []int64) error {
	m.reorderCredCalls = append(m.reorderCredCalls, orderedIDs)
	return m.err
}

func (m *mockAcademicStore) ListCertifications(_ context.Context) ([]domain.Certification, error) {
	return m.certs, m.err
}

func (m *mockAcademicStore) CreateCertification(_ context.Context, input domain.CertificationInput) (domain.Certification, error) {
	m.createCertCalls = append(m.createCertCalls, input)
	if m.err != nil {
		return domain.Certification{}, m.err
	}
	return m.cert, nil
}

func (m *mockAcademicStore) UpdateCertification(_ context.Context, id int64, input domain.CertificationInput) (domain.Certification, error) {
	m.updateCertCalls = append(m.updateCertCalls, updateCertCall{ID: id, Input: input})
	if m.err != nil {
		return domain.Certification{}, m.err
	}
	return m.cert, nil
}

func (m *mockAcademicStore) DeleteCertification(_ context.Context, id int64) error {
	m.deleteCertCalls = append(m.deleteCertCalls, id)
	return m.err
}

func (m *mockAcademicStore) ReorderCertifications(_ context.Context, orderedIDs []int64) error {
	m.reorderCertCalls = append(m.reorderCertCalls, orderedIDs)
	return m.err
}

// =================================================================
// Academic Credentials — ListAcademicCredentials
// =================================================================

func TestAcademicService_ListAcademicCredentials_Success(t *testing.T) {
	store := &mockAcademicStore{
		credentials: []domain.AcademicCredential{
			{ID: 1, Institution: "MIT", CredentialType: "B.S.", FieldOfStudy: "CS"},
		},
	}
	svc := NewAcademicService(store)

	creds, err := svc.ListAcademicCredentials(context.Background())
	require.NoError(t, err)
	require.Len(t, creds, 1)
	assert.Equal(t, "MIT", creds[0].Institution)
}

func TestAcademicService_ListAcademicCredentials_Empty(t *testing.T) {
	store := &mockAcademicStore{credentials: []domain.AcademicCredential{}}
	svc := NewAcademicService(store)

	creds, err := svc.ListAcademicCredentials(context.Background())
	require.NoError(t, err)
	assert.Empty(t, creds)
}

func TestAcademicService_ListAcademicCredentials_StoreError(t *testing.T) {
	store := &mockAcademicStore{err: fmt.Errorf("db error")}
	svc := NewAcademicService(store)

	_, err := svc.ListAcademicCredentials(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// =================================================================
// CreateAcademicCredential — validation
// =================================================================

func TestAcademicService_CreateAcademicCredential_EmptyInstitution(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:    "",
		CredentialType: "B.S.",
		FieldOfStudy:   "CS",
		CompletionDate: "2020",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "institution")
	assert.Empty(t, store.createCredCalls)
}

func TestAcademicService_CreateAcademicCredential_EmptyCredentialType(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:    "MIT",
		CredentialType: "",
		FieldOfStudy:   "CS",
		CompletionDate: "2020",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credential type")
	assert.Empty(t, store.createCredCalls)
}

func TestAcademicService_CreateAcademicCredential_EmptyFieldOfStudy(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:    "MIT",
		CredentialType: "B.S.",
		FieldOfStudy:   "",
		CompletionDate: "2020",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field of study")
	assert.Empty(t, store.createCredCalls)
}

func TestAcademicService_CreateAcademicCredential_EmptyCompletionDate(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:    "MIT",
		CredentialType: "B.S.",
		FieldOfStudy:   "CS",
		CompletionDate: "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "completion date")
	assert.Empty(t, store.createCredCalls)
}

func TestAcademicService_CreateAcademicCredential_InvalidGranularity(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "B.S.",
		FieldOfStudy:    "CS",
		CompletionDate:  "2020",
		DateGranularity: "invalid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "granularity")
	assert.Empty(t, store.createCredCalls)
}

// =================================================================
// CreateAcademicCredential — happy path
// =================================================================

func TestAcademicService_CreateAcademicCredential_Success(t *testing.T) {
	store := &mockAcademicStore{
		credential: domain.AcademicCredential{
			ID: 1, Institution: "MIT", CredentialType: "B.S.",
			FieldOfStudy: "Computer Science", CompletionDate: "2020",
		},
	}
	svc := NewAcademicService(store)

	cred, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:    "MIT",
		CredentialType: "B.S.",
		FieldOfStudy:   "Computer Science",
		CompletionDate: "2020",
	})
	require.NoError(t, err)
	assert.Equal(t, "MIT", cred.Institution)
	require.Len(t, store.createCredCalls, 1)
}

func TestAcademicService_CreateAcademicCredential_WithGranularity(t *testing.T) {
	store := &mockAcademicStore{
		credential: domain.AcademicCredential{
			ID: 1, Institution: "MIT", DateGranularity: "month",
		},
	}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "B.S.",
		FieldOfStudy:    "CS",
		CompletionDate:  "2020-06",
		DateGranularity: "month",
	})
	require.NoError(t, err)
}

func TestAcademicService_CreateAcademicCredential_StoreError(t *testing.T) {
	store := &mockAcademicStore{err: fmt.Errorf("db full")}
	svc := NewAcademicService(store)

	_, err := svc.CreateAcademicCredential(context.Background(), domain.AcademicInput{
		Institution:    "MIT",
		CredentialType: "B.S.",
		FieldOfStudy:   "CS",
		CompletionDate: "2020",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db full")
}

// =================================================================
// UpdateAcademicCredential — validation
// =================================================================

func TestAcademicService_UpdateAcademicCredential_EmptyInstitution(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.UpdateAcademicCredential(context.Background(), 1, domain.AcademicInput{
		Institution:    "",
		CredentialType: "B.S.",
		FieldOfStudy:   "CS",
		CompletionDate: "2020",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "institution")
	assert.Empty(t, store.updateCredCalls)
}

// =================================================================
// UpdateAcademicCredential — happy path
// =================================================================

func TestAcademicService_UpdateAcademicCredential_Success(t *testing.T) {
	store := &mockAcademicStore{
		credential: domain.AcademicCredential{
			ID: 1, Institution: "Stanford",
		},
	}
	svc := NewAcademicService(store)

	cred, err := svc.UpdateAcademicCredential(context.Background(), 1, domain.AcademicInput{
		Institution:    "Stanford",
		CredentialType: "M.S.",
		FieldOfStudy:   "CS",
		CompletionDate: "2022",
	})
	require.NoError(t, err)
	assert.Equal(t, "Stanford", cred.Institution)
	require.Len(t, store.updateCredCalls, 1)
	assert.Equal(t, int64(1), store.updateCredCalls[0].ID)
}

// =================================================================
// DeleteAcademicCredential
// =================================================================

func TestAcademicService_DeleteAcademicCredential_Success(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	err := svc.DeleteAcademicCredential(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, store.deleteCredCalls, 1)
	assert.Equal(t, int64(1), store.deleteCredCalls[0])
}

func TestAcademicService_DeleteAcademicCredential_StoreError(t *testing.T) {
	store := &mockAcademicStore{err: fmt.Errorf("not found")}
	svc := NewAcademicService(store)

	err := svc.DeleteAcademicCredential(context.Background(), 999)
	require.Error(t, err)
}

// =================================================================
// ReorderAcademicCredentials
// =================================================================

func TestAcademicService_ReorderAcademicCredentials_Success(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	err := svc.ReorderAcademicCredentials(context.Background(), []int64{3, 1, 2})
	require.NoError(t, err)
	require.Len(t, store.reorderCredCalls, 1)
	assert.Equal(t, []int64{3, 1, 2}, store.reorderCredCalls[0])
}

func TestAcademicService_ReorderAcademicCredentials_Empty(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	err := svc.ReorderAcademicCredentials(context.Background(), []int64{})
	require.NoError(t, err)
}

// =================================================================
// Certifications — ListCertifications
// =================================================================

func TestAcademicService_ListCertifications_Success(t *testing.T) {
	store := &mockAcademicStore{
		certs: []domain.Certification{
			{ID: 1, Name: "AWS SAA", IssuingBody: "Amazon", IsActive: true},
		},
	}
	svc := NewAcademicService(store)

	certs, err := svc.ListCertifications(context.Background())
	require.NoError(t, err)
	require.Len(t, certs, 1)
	assert.Equal(t, "AWS SAA", certs[0].Name)
	assert.True(t, certs[0].IsActive)
}

func TestAcademicService_ListCertifications_Empty(t *testing.T) {
	store := &mockAcademicStore{certs: []domain.Certification{}}
	svc := NewAcademicService(store)

	certs, err := svc.ListCertifications(context.Background())
	require.NoError(t, err)
	assert.Empty(t, certs)
}

// =================================================================
// CreateCertification — validation
// =================================================================

func TestAcademicService_CreateCertification_EmptyName(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateCertification(context.Background(), domain.CertificationInput{
		Name:        "",
		IssuingBody: "Amazon",
		DateEarned:  "2023-01",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "certification name")
	assert.Empty(t, store.createCertCalls)
}

func TestAcademicService_CreateCertification_EmptyIssuingBody(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateCertification(context.Background(), domain.CertificationInput{
		Name:        "AWS SAA",
		IssuingBody: "",
		DateEarned:  "2023-01",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issuing body")
	assert.Empty(t, store.createCertCalls)
}

func TestAcademicService_CreateCertification_EmptyDateEarned(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.CreateCertification(context.Background(), domain.CertificationInput{
		Name:        "AWS SAA",
		IssuingBody: "Amazon",
		DateEarned:  "",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "date earned")
	assert.Empty(t, store.createCertCalls)
}

// =================================================================
// CreateCertification — happy path
// =================================================================

func TestAcademicService_CreateCertification_Success(t *testing.T) {
	store := &mockAcademicStore{
		cert: domain.Certification{
			ID: 1, Name: "AWS SAA", IssuingBody: "Amazon",
			DateEarned: "2023-01", IsActive: true,
		},
	}
	svc := NewAcademicService(store)

	cert, err := svc.CreateCertification(context.Background(), domain.CertificationInput{
		Name:        "AWS SAA",
		IssuingBody: "Amazon",
		DateEarned:  "2023-01",
	})
	require.NoError(t, err)
	assert.Equal(t, "AWS SAA", cert.Name)
	require.Len(t, store.createCertCalls, 1)
}

func TestAcademicService_CreateCertification_WithExpiration(t *testing.T) {
	store := &mockAcademicStore{
		cert: domain.Certification{
			ID: 1, Name: "AWS SAA", ExpirationDate: "2026-01", IsActive: true,
		},
	}
	svc := NewAcademicService(store)

	cert, err := svc.CreateCertification(context.Background(), domain.CertificationInput{
		Name:           "AWS SAA",
		IssuingBody:    "Amazon",
		DateEarned:     "2023-01",
		ExpirationDate: "2026-01",
	})
	require.NoError(t, err)
	assert.Equal(t, "2026-01", cert.ExpirationDate)
}

func TestAcademicService_CreateCertification_StoreError(t *testing.T) {
	store := &mockAcademicStore{err: fmt.Errorf("db full")}
	svc := NewAcademicService(store)

	_, err := svc.CreateCertification(context.Background(), domain.CertificationInput{
		Name:        "AWS SAA",
		IssuingBody: "Amazon",
		DateEarned:  "2023-01",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db full")
}

// =================================================================
// UpdateCertification — validation
// =================================================================

func TestAcademicService_UpdateCertification_EmptyName(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	_, err := svc.UpdateCertification(context.Background(), 1, domain.CertificationInput{
		Name:        "",
		IssuingBody: "Amazon",
		DateEarned:  "2023-01",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "certification name")
	assert.Empty(t, store.updateCertCalls)
}

// =================================================================
// UpdateCertification — happy path
// =================================================================

func TestAcademicService_UpdateCertification_Success(t *testing.T) {
	store := &mockAcademicStore{
		cert: domain.Certification{
			ID: 1, Name: "AWS SAP", IssuingBody: "Amazon",
		},
	}
	svc := NewAcademicService(store)

	cert, err := svc.UpdateCertification(context.Background(), 1, domain.CertificationInput{
		Name:        "AWS SAP",
		IssuingBody: "Amazon",
		DateEarned:  "2023-06",
	})
	require.NoError(t, err)
	assert.Equal(t, "AWS SAP", cert.Name)
	require.Len(t, store.updateCertCalls, 1)
	assert.Equal(t, int64(1), store.updateCertCalls[0].ID)
}

// =================================================================
// DeleteCertification
// =================================================================

func TestAcademicService_DeleteCertification_Success(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	err := svc.DeleteCertification(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, store.deleteCertCalls, 1)
	assert.Equal(t, int64(1), store.deleteCertCalls[0])
}

func TestAcademicService_DeleteCertification_StoreError(t *testing.T) {
	store := &mockAcademicStore{err: fmt.Errorf("not found")}
	svc := NewAcademicService(store)

	err := svc.DeleteCertification(context.Background(), 999)
	require.Error(t, err)
}

// =================================================================
// ReorderCertifications
// =================================================================

func TestAcademicService_ReorderCertifications_Success(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	err := svc.ReorderCertifications(context.Background(), []int64{2, 3, 1})
	require.NoError(t, err)
	require.Len(t, store.reorderCertCalls, 1)
	assert.Equal(t, []int64{2, 3, 1}, store.reorderCertCalls[0])
}

func TestAcademicService_ReorderCertifications_Empty(t *testing.T) {
	store := &mockAcademicStore{}
	svc := NewAcademicService(store)

	err := svc.ReorderCertifications(context.Background(), []int64{})
	require.NoError(t, err)
}
