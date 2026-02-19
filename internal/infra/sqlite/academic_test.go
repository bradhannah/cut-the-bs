package sqlite

import (
	"context"
	"testing"

	"cut-the-bs/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =================================================================
// Academic Credentials
// =================================================================

func TestCreateAcademicCredential(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cred, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "degree",
		FieldOfStudy:    "Computer Science",
		CompletionDate:  "2020",
		DateGranularity: "year",
	})
	require.NoError(t, err)
	assert.NotZero(t, cred.ID)
	assert.Equal(t, "MIT", cred.Institution)
	assert.Equal(t, "degree", cred.CredentialType)
	assert.Equal(t, "Computer Science", cred.FieldOfStudy)
	assert.Equal(t, "2020", cred.CompletionDate)
	assert.Equal(t, "year", cred.DateGranularity)
	assert.Equal(t, 1, cred.SortOrder)
	assert.NotEmpty(t, cred.CreatedAt)
}

func TestCreateAcademicCredential_AutoIncrementsSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "degree",
		FieldOfStudy:    "CS",
		CompletionDate:  "2020",
		DateGranularity: "year",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, first.SortOrder)

	second, err := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution:     "Stanford",
		CredentialType:  "degree",
		FieldOfStudy:    "Math",
		CompletionDate:  "2022",
		DateGranularity: "year",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, second.SortOrder)
}

func TestListAcademicCredentials_OrderedBySortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2020", DateGranularity: "year",
	})
	_, _ = store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "Stanford", CredentialType: "degree",
		FieldOfStudy: "Math", CompletionDate: "2022", DateGranularity: "year",
	})

	creds, err := store.ListAcademicCredentials(ctx)
	require.NoError(t, err)
	require.Len(t, creds, 2)
	assert.Equal(t, "MIT", creds[0].Institution)
	assert.Equal(t, "Stanford", creds[1].Institution)
}

func TestListAcademicCredentials_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	creds, err := store.ListAcademicCredentials(ctx)
	require.NoError(t, err)
	assert.NotNil(t, creds)
	assert.Len(t, creds, 0)
}

func TestUpdateAcademicCredential(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cred, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2020", DateGranularity: "year",
	})

	updated, err := store.UpdateAcademicCredential(ctx, cred.ID, domain.AcademicInput{
		Institution:     "MIT",
		CredentialType:  "masters",
		FieldOfStudy:    "AI",
		CompletionDate:  "2022-06",
		DateGranularity: "month",
	})
	require.NoError(t, err)
	assert.Equal(t, cred.ID, updated.ID)
	assert.Equal(t, "masters", updated.CredentialType)
	assert.Equal(t, "AI", updated.FieldOfStudy)
	assert.Equal(t, "2022-06", updated.CompletionDate)
	assert.Equal(t, "month", updated.DateGranularity)
}

func TestUpdateAcademicCredential_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateAcademicCredential(ctx, 999, domain.AcademicInput{
		Institution: "MIT", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2020", DateGranularity: "year",
	})
	require.Error(t, err)
}

func TestDeleteAcademicCredential(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cred, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "MIT", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2020", DateGranularity: "year",
	})

	err := store.DeleteAcademicCredential(ctx, cred.ID)
	require.NoError(t, err)

	creds, err := store.ListAcademicCredentials(ctx)
	require.NoError(t, err)
	assert.Len(t, creds, 0)
}

func TestDeleteAcademicCredential_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteAcademicCredential(ctx, 999)
	require.Error(t, err)
}

func TestReorderAcademicCredentials(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	a, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "A", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2020", DateGranularity: "year",
	})
	b, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "B", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2021", DateGranularity: "year",
	})
	c, _ := store.CreateAcademicCredential(ctx, domain.AcademicInput{
		Institution: "C", CredentialType: "degree",
		FieldOfStudy: "CS", CompletionDate: "2022", DateGranularity: "year",
	})

	err := store.ReorderAcademicCredentials(ctx, []int64{c.ID, b.ID, a.ID})
	require.NoError(t, err)

	creds, err := store.ListAcademicCredentials(ctx)
	require.NoError(t, err)
	require.Len(t, creds, 3)
	assert.Equal(t, "C", creds[0].Institution)
	assert.Equal(t, "B", creds[1].Institution)
	assert.Equal(t, "A", creds[2].Institution)
}

// =================================================================
// Certifications
// =================================================================

func TestCreateCertification(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, err := store.CreateCertification(ctx, domain.CertificationInput{
		Name:        "AWS SAA",
		IssuingBody: "Amazon",
		DateEarned:  "2023-01-15",
	})
	require.NoError(t, err)
	assert.NotZero(t, cert.ID)
	assert.Equal(t, "AWS SAA", cert.Name)
	assert.Equal(t, "Amazon", cert.IssuingBody)
	assert.Equal(t, "2023-01-15", cert.DateEarned)
	assert.Empty(t, cert.ExpirationDate)
	assert.True(t, cert.IsActive, "cert with no expiration should be active")
	assert.Equal(t, 1, cert.SortOrder)
}

func TestCreateCertification_WithExpiration(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, err := store.CreateCertification(ctx, domain.CertificationInput{
		Name:           "AWS SAA",
		IssuingBody:    "Amazon",
		DateEarned:     "2023-01-15",
		ExpirationDate: "2099-01-15",
	})
	require.NoError(t, err)
	assert.Equal(t, "2099-01-15", cert.ExpirationDate)
	assert.True(t, cert.IsActive, "cert with future expiration should be active")
}

func TestCreateCertification_Expired(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, err := store.CreateCertification(ctx, domain.CertificationInput{
		Name:           "Old Cert",
		IssuingBody:    "OldCo",
		DateEarned:     "2010-01-01",
		ExpirationDate: "2015-01-01",
	})
	require.NoError(t, err)
	assert.False(t, cert.IsActive, "cert with past expiration should be inactive")
}

func TestCreateCertification_AutoIncrementsSortOrder(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	first, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "A", IssuingBody: "X", DateEarned: "2023-01-01",
	})
	second, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "B", IssuingBody: "Y", DateEarned: "2023-06-01",
	})
	assert.Equal(t, 1, first.SortOrder)
	assert.Equal(t, 2, second.SortOrder)
}

func TestListCertifications(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, _ = store.CreateCertification(ctx, domain.CertificationInput{
		Name: "A", IssuingBody: "X", DateEarned: "2023-01-01",
	})
	_, _ = store.CreateCertification(ctx, domain.CertificationInput{
		Name: "B", IssuingBody: "Y", DateEarned: "2023-06-01",
	})

	certs, err := store.ListCertifications(ctx)
	require.NoError(t, err)
	require.Len(t, certs, 2)
	assert.Equal(t, "A", certs[0].Name)
	assert.Equal(t, "B", certs[1].Name)
}

func TestListCertifications_EmptyReturnsEmptySlice(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	certs, err := store.ListCertifications(ctx)
	require.NoError(t, err)
	assert.NotNil(t, certs)
	assert.Len(t, certs, 0)
}

func TestUpdateCertification(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "AWS SAA", IssuingBody: "Amazon", DateEarned: "2023-01-15",
	})

	updated, err := store.UpdateCertification(ctx, cert.ID, domain.CertificationInput{
		Name:           "AWS SAP",
		IssuingBody:    "Amazon Web Services",
		DateEarned:     "2023-06-01",
		ExpirationDate: "2099-06-01",
	})
	require.NoError(t, err)
	assert.Equal(t, cert.ID, updated.ID)
	assert.Equal(t, "AWS SAP", updated.Name)
	assert.Equal(t, "Amazon Web Services", updated.IssuingBody)
	assert.True(t, updated.IsActive)
}

func TestUpdateCertification_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	_, err := store.UpdateCertification(ctx, 999, domain.CertificationInput{
		Name: "X", IssuingBody: "Y", DateEarned: "2023-01-01",
	})
	require.Error(t, err)
}

func TestDeleteCertification(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	cert, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "A", IssuingBody: "X", DateEarned: "2023-01-01",
	})

	err := store.DeleteCertification(ctx, cert.ID)
	require.NoError(t, err)

	certs, err := store.ListCertifications(ctx)
	require.NoError(t, err)
	assert.Len(t, certs, 0)
}

func TestDeleteCertification_NotFound(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	err := store.DeleteCertification(ctx, 999)
	require.Error(t, err)
}

func TestReorderCertifications(t *testing.T) {
	store := testStore(t)
	require.NoError(t, Migrate(store))
	ctx := context.Background()

	a, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "A", IssuingBody: "X", DateEarned: "2023-01-01",
	})
	b, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "B", IssuingBody: "Y", DateEarned: "2023-06-01",
	})
	c, _ := store.CreateCertification(ctx, domain.CertificationInput{
		Name: "C", IssuingBody: "Z", DateEarned: "2023-12-01",
	})

	err := store.ReorderCertifications(ctx, []int64{c.ID, b.ID, a.ID})
	require.NoError(t, err)

	certs, err := store.ListCertifications(ctx)
	require.NoError(t, err)
	require.Len(t, certs, 3)
	assert.Equal(t, "C", certs[0].Name)
	assert.Equal(t, "B", certs[1].Name)
	assert.Equal(t, "A", certs[2].Name)
}
