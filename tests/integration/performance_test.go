package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cut-the-bs/internal/domain"
	"cut-the-bs/internal/infra/pdf"
	"cut-the-bs/internal/infra/sqlite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// perfStore creates a store and seeds it with realistic data volume.
func perfStore(t *testing.T) *sqlite.Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "perf.db")
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	store, err := sqlite.NewStore(dbPath, logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, sqlite.Migrate(store))
	return store
}

// seedFullResume populates the store with a realistic resume's worth
// of data: profile, links, multiple work history entries with bullets,
// skills across categories, education, certifications, summaries,
// and descriptors.
func seedFullResume(t *testing.T, store *sqlite.Store) domain.RenderResumeRequest {
	t.Helper()
	ctx := context.Background()

	_, err := store.GetProfile(ctx) // ensure row exists
	require.NoError(t, err)
	profile, err := store.UpdateProfile(ctx, domain.UserProfile{
		FullName: "Performance Test User",
		Email:    "perf@example.com",
		Phone:    "555-9999",
		Location: "Seattle, WA",
	})
	require.NoError(t, err)

	var links []domain.ProfileLink
	for _, l := range []struct{ label, url string }{
		{"LinkedIn", "https://linkedin.com/in/perfuser"},
		{"GitHub", "https://github.com/perfuser"},
		{"Portfolio", "https://perfuser.dev"},
	} {
		link, err := store.CreateProfileLink(ctx, domain.ProfileLinkInput{
			Label: l.label,
			URL:   l.url,
		})
		require.NoError(t, err)
		links = append(links, link)
	}

	summary, err := store.CreateSummary(ctx, domain.SummaryInput{
		Label:    "General",
		BodyText: "Senior software engineer with 12+ years of experience in distributed systems, cloud architecture, and technical leadership. Proven track record of designing scalable solutions processing millions of requests daily across multiple cloud providers.",
	})
	require.NoError(t, err)

	// Create 5 work history entries with 3-5 bullets each.
	employers := []struct {
		name, title, start, end string
		bullets                 []string
	}{
		{
			"MegaCorp", "Principal Engineer", "2022-01", "",
			[]string{
				"Architected event-driven microservices platform handling 50M daily transactions across 12 regions",
				"Led team of 15 engineers in migration from monolith to service mesh architecture",
				"Reduced cloud infrastructure costs by 35% through auto-scaling optimization and spot instance strategies",
				"Established engineering standards adopted by 200+ developers across the organization",
			},
		},
		{
			"CloudScale Inc", "Staff Software Engineer", "2019-06", "2021-12",
			[]string{
				"Designed real-time data streaming pipeline processing 10TB daily with sub-second latency",
				"Built multi-tenant SaaS platform serving 5000+ enterprise customers",
				"Mentored 8 engineers, with 3 promoted to senior positions within 18 months",
				"Implemented zero-downtime deployment strategy reducing release risk by 90%",
			},
		},
		{
			"DataFlow Systems", "Senior Software Engineer", "2016-03", "2019-05",
			[]string{
				"Developed distributed caching layer reducing database load by 60%",
				"Created CI/CD pipeline reducing deployment time from 2 hours to 15 minutes",
				"Led security audit resulting in SOC 2 Type II certification",
			},
		},
		{
			"WebTech Solutions", "Software Engineer", "2013-08", "2016-02",
			[]string{
				"Built RESTful APIs serving 100K daily active users with 99.99% uptime",
				"Implemented OAuth 2.0 authentication system used by 50+ partner integrations",
				"Optimized database queries reducing average response time from 200ms to 20ms",
				"Developed automated testing framework increasing code coverage from 40% to 92%",
				"Contributed to open-source projects used by 10,000+ developers worldwide",
			},
		},
		{
			"StartupLabs", "Junior Developer", "2011-06", "2013-07",
			[]string{
				"Built customer-facing web application using React and Node.js",
				"Implemented payment processing integration handling $2M monthly transactions",
				"Created automated reporting system reducing manual effort by 20 hours per week",
			},
		},
	}

	var workHistory []domain.WorkHistoryEntry
	for _, emp := range employers {
		granEnd := ""
		if emp.end != "" {
			granEnd = "month"
		}
		wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
			EmployerName:         emp.name,
			JobTitle:             emp.title,
			StartDate:            emp.start,
			EndDate:              emp.end,
			DateGranularityStart: "month",
			DateGranularityEnd:   granEnd,
		})
		require.NoError(t, err)

		var bullets []domain.AchievementBullet
		for _, bt := range emp.bullets {
			b, err := store.CreateBullet(ctx, wh.ID, bt, domain.BulletTypePrimary)
			require.NoError(t, err)
			bullets = append(bullets, b)
		}
		wh.Bullets = bullets
		workHistory = append(workHistory, wh)
	}

	// Create 4 skill categories with 3-6 skills each.
	skillCategories := []struct {
		name   string
		skills []struct {
			name  string
			level int
		}
	}{
		{"Programming Languages", []struct {
			name  string
			level int
		}{
			{"Go", 10}, {"Python", 9}, {"TypeScript", 8},
			{"Java", 7}, {"Rust", 6}, {"C++", 5},
		}},
		{"Cloud & Infrastructure", []struct {
			name  string
			level int
		}{
			{"AWS", 9}, {"GCP", 8}, {"Kubernetes", 9},
			{"Terraform", 8}, {"Docker", 9},
		}},
		{"Databases", []struct {
			name  string
			level int
		}{
			{"PostgreSQL", 9}, {"Redis", 8}, {"MongoDB", 7},
			{"DynamoDB", 7},
		}},
		{"Frameworks & Tools", []struct {
			name  string
			level int
		}{
			{"gRPC", 8}, {"GraphQL", 7}, {"React", 7},
			{"Kafka", 8}, {"Git", 10},
		}},
	}

	var allSkills []domain.Skill
	for _, sc := range skillCategories {
		cat, err := store.CreateSkillCategory(ctx, sc.name)
		require.NoError(t, err)
		for _, sk := range sc.skills {
			skill, err := store.CreateSkill(ctx, domain.SkillInput{
				Name:            sk.name,
				CategoryID:      cat.ID,
				CompetenceLevel: sk.level,
			})
			require.NoError(t, err)
			allSkills = append(allSkills, skill)
		}
	}

	// Education and certifications.
	academics := []domain.AcademicInput{
		{Institution: "Stanford University", CredentialType: "Master of Science", FieldOfStudy: "Computer Science", CompletionDate: "2011-06", DateGranularity: "month"},
		{Institution: "UC Berkeley", CredentialType: "Bachelor of Science", FieldOfStudy: "Electrical Engineering", CompletionDate: "2009-05", DateGranularity: "month"},
	}

	var acadList []domain.AcademicCredential
	for _, a := range academics {
		acad, err := store.CreateAcademicCredential(ctx, a)
		require.NoError(t, err)
		acadList = append(acadList, acad)
	}

	certInputs := []domain.CertificationInput{
		{Name: "AWS Solutions Architect Professional", IssuingBody: "Amazon Web Services", DateEarned: "2023-06"},
		{Name: "GCP Professional Cloud Architect", IssuingBody: "Google Cloud", DateEarned: "2023-01"},
		{Name: "Certified Kubernetes Administrator", IssuingBody: "CNCF", DateEarned: "2022-08"},
	}

	var certList []domain.Certification
	for _, c := range certInputs {
		cert, err := store.CreateCertification(ctx, c)
		require.NoError(t, err)
		certList = append(certList, cert)
	}

	descriptors := []string{
		"Principal Software Engineer",
		"Technical Lead",
		"Distributed Systems Architect",
	}

	var descList []domain.RoleDescriptor
	for _, d := range descriptors {
		desc, err := store.CreateDescriptor(ctx, d)
		require.NoError(t, err)
		descList = append(descList, desc)
	}

	profTmpl := pdf.ProfessionalTemplate()
	return domain.RenderResumeRequest{
		Template:    &profTmpl,
		OutputDir:   t.TempDir(),
		Profile:     profile,
		Links:       links,
		Summaries:   []domain.ProfessionalSummary{summary},
		WorkHistory: workHistory,
		Skills:      allSkills,
		Academics:   acadList,
		Certs:       certList,
		Descriptors: descList,
	}
}

func TestPerformance_PDFGenerationUnder5Seconds(t *testing.T) {
	store := perfStore(t)
	req := seedFullResume(t, store)
	renderer := pdf.NewRenderer()

	// Run PDF generation multiple times to get a reliable measurement.
	const iterations = 3
	for i := range iterations {
		req.OutputDir = t.TempDir()
		start := time.Now()

		path, err := renderer.RenderResume(context.Background(), req)
		elapsed := time.Since(start)

		require.NoError(t, err, "iteration %d", i)
		assert.FileExists(t, path, "iteration %d", i)
		assert.Less(t, elapsed, 5*time.Second,
			"PDF generation should complete within 5 seconds (took %v on iteration %d)",
			elapsed, i)

		t.Logf("PDF generation iteration %d: %v", i, elapsed)
	}
}

func TestPerformance_PDFBothTemplatesUnder5Seconds(t *testing.T) {
	store := perfStore(t)
	req := seedFullResume(t, store)
	renderer := pdf.NewRenderer()

	templates := map[string]domain.TemplateDetail{
		"professional": pdf.ProfessionalTemplate(),
		"modern":       pdf.ModernTemplate(),
	}
	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			req.Template = &tmpl
			req.OutputDir = t.TempDir()

			start := time.Now()
			path, err := renderer.RenderResume(context.Background(), req)
			elapsed := time.Since(start)

			require.NoError(t, err)
			assert.FileExists(t, path)
			assert.Less(t, elapsed, 5*time.Second,
				"template %q should render within 5 seconds (took %v)", name, elapsed)

			t.Logf("Template %q: %v", name, elapsed)
		})
	}
}

func TestPerformance_DatabaseOperationsUnder100ms(t *testing.T) {
	store := perfStore(t)
	ctx := context.Background()

	// Seed some data first so queries have something to work with.
	_ = seedFullResume(t, store)

	// Measure common database read operations.
	operations := []struct {
		name string
		fn   func() error
	}{
		{"GetProfile", func() error {
			_, err := store.GetProfile(ctx)
			return err
		}},
		{"ListProfileLinks", func() error {
			_, err := store.ListProfileLinks(ctx)
			return err
		}},
		{"ListWorkHistory", func() error {
			_, err := store.ListWorkHistory(ctx)
			return err
		}},
		{"ListSkillCategories", func() error {
			_, err := store.ListSkillCategories(ctx)
			return err
		}},
		{"ListAcademicCredentials", func() error {
			_, err := store.ListAcademicCredentials(ctx)
			return err
		}},
		{"ListCertifications", func() error {
			_, err := store.ListCertifications(ctx)
			return err
		}},
		{"ListSummaries", func() error {
			_, err := store.ListSummaries(ctx)
			return err
		}},
		{"ListDescriptors", func() error {
			_, err := store.ListDescriptors(ctx)
			return err
		}},
		{"ExportAllData", func() error {
			_, err := store.ExportAllData(ctx)
			return err
		}},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			start := time.Now()
			err := op.fn()
			elapsed := time.Since(start)

			require.NoError(t, err)
			assert.Less(t, elapsed, 100*time.Millisecond,
				"%s should complete within 100ms (took %v)", op.name, elapsed)

			t.Logf("%s: %v", op.name, elapsed)
		})
	}
}

func TestPerformance_DatabaseWriteOperationsUnder100ms(t *testing.T) {
	store := perfStore(t)
	ctx := context.Background()

	// Seed prerequisite data.
	_, err := store.GetProfile(ctx) // ensure profile row exists
	require.NoError(t, err)

	cat, err := store.CreateSkillCategory(ctx, "PerfTest")
	require.NoError(t, err)

	wh, err := store.CreateWorkHistory(ctx, domain.WorkHistoryInput{
		EmployerName:         "PerfTest Corp",
		JobTitle:             "Engineer",
		StartDate:            "2020-01",
		DateGranularityStart: "month",
	})
	require.NoError(t, err)

	writeOps := []struct {
		name string
		fn   func() error
	}{
		{"CreateSkill", func() error {
			_, err := store.CreateSkill(ctx, domain.SkillInput{
				Name:       fmt.Sprintf("Skill-%d", time.Now().UnixNano()),
				CategoryID: cat.ID, CompetenceLevel: 5,
			})
			return err
		}},
		{"CreateBullet", func() error {
			_, err := store.CreateBullet(ctx, wh.ID, fmt.Sprintf("Bullet-%d", time.Now().UnixNano()), domain.BulletTypePrimary)
			return err
		}},
		{"UpdateProfile", func() error {
			_, err := store.UpdateProfile(ctx, domain.UserProfile{
				FullName: "Updated User",
				Email:    "updated@example.com",
				Phone:    "555-0000",
				Location: "Anywhere, USA",
			})
			return err
		}},
		{"CreateSummary", func() error {
			_, err := store.CreateSummary(ctx, domain.SummaryInput{
				Label:    fmt.Sprintf("Summary-%d", time.Now().UnixNano()),
				BodyText: "Performance test summary text.",
			})
			return err
		}},
		{"CreateDescriptor", func() error {
			_, err := store.CreateDescriptor(ctx, fmt.Sprintf("Descriptor-%d", time.Now().UnixNano()))
			return err
		}},
	}

	for _, op := range writeOps {
		t.Run(op.name, func(t *testing.T) {
			start := time.Now()
			err := op.fn()
			elapsed := time.Since(start)

			require.NoError(t, err)
			assert.Less(t, elapsed, 100*time.Millisecond,
				"%s should complete within 100ms (took %v)", op.name, elapsed)

			t.Logf("%s: %v", op.name, elapsed)
		})
	}
}

func TestPerformance_ExportImportRoundtripReasonable(t *testing.T) {
	store := perfStore(t)
	ctx := context.Background()

	// Seed full resume data.
	_ = seedFullResume(t, store)

	// Measure export time.
	exportDir := t.TempDir()
	exportPath := filepath.Join(exportDir, "perf-export.json")

	start := time.Now()
	data, err := store.ExportAllData(ctx)
	exportElapsed := time.Since(start)
	require.NoError(t, err)
	assert.NotEmpty(t, data.Profile.FullName)
	assert.Less(t, exportElapsed, 1*time.Second,
		"ExportAllData should complete within 1 second (took %v)", exportElapsed)
	t.Logf("ExportAllData: %v", exportElapsed)

	// Write JSON (for import test).
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(exportPath, jsonBytes, 0o644))

	// Measure import time into a fresh store.
	store2 := perfStore(t)
	start = time.Now()
	err = store2.ImportAllData(ctx, data)
	importElapsed := time.Since(start)
	require.NoError(t, err)
	assert.Less(t, importElapsed, 2*time.Second,
		"ImportAllData should complete within 2 seconds (took %v)", importElapsed)
	t.Logf("ImportAllData: %v", importElapsed)
}
