package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Date validation tests ---

func TestValidateDateRange_ValidSameGranularity(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		startGran string
		end       string
		endGran   string
	}{
		{"year same", "2023", "year", "2023", "year"},
		{"year end after", "2022", "year", "2023", "year"},
		{"month same", "2023-03", "month", "2023-03", "month"},
		{"month end after", "2023-01", "month", "2023-06", "month"},
		{"day same", "2023-03-15", "day", "2023-03-15", "day"},
		{"day end after", "2023-01-01", "day", "2023-12-31", "day"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.start, tt.startGran, tt.end, tt.endGran)
			assert.NoError(t, err)
		})
	}
}

func TestValidateDateRange_ValidMixedGranularity(t *testing.T) {
	// When granularities differ, compare at the coarsest level.
	tests := []struct {
		name      string
		start     string
		startGran string
		end       string
		endGran   string
	}{
		{"year vs month, end after", "2022", "year", "2023-06", "month"},
		{"month vs day, end after", "2023-01", "month", "2023-06-15", "day"},
		{"year vs day, same year", "2023", "year", "2023-01-01", "day"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.start, tt.startGran, tt.end, tt.endGran)
			assert.NoError(t, err)
		})
	}
}

func TestValidateDateRange_EmptyEndIsPresent(t *testing.T) {
	// Empty end date means "present" — always valid.
	err := ValidateDateRange("2023", "year", "", "")
	assert.NoError(t, err)
}

func TestValidateDateRange_EndBeforeStart(t *testing.T) {
	tests := []struct {
		name      string
		start     string
		startGran string
		end       string
		endGran   string
	}{
		{"year", "2024", "year", "2023", "year"},
		{"month", "2023-06", "month", "2023-01", "month"},
		{"day", "2023-06-15", "day", "2023-01-01", "day"},
		{"mixed: year vs month", "2024", "year", "2023-12", "month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.start, tt.startGran, tt.end, tt.endGran)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "end date")
		})
	}
}

// --- Required field tests ---

func TestValidateRequired_NonEmpty(t *testing.T) {
	err := ValidateRequired("hello", "name")
	assert.NoError(t, err)
}

func TestValidateRequired_Empty(t *testing.T) {
	err := ValidateRequired("", "name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidateRequired_Whitespace(t *testing.T) {
	err := ValidateRequired("   ", "name")
	assert.Error(t, err, "whitespace-only should be treated as empty")
}

// --- Email validation tests ---

func TestValidateEmail_Valid(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"first.last@company.org",
		"name+tag@domain.co.uk",
	}
	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			err := ValidateEmail(email)
			assert.NoError(t, err)
		})
	}
}

func TestValidateEmail_Invalid(t *testing.T) {
	invalidEmails := []string{
		"",
		"notanemail",
		"@domain.com",
		"user@",
		"user@.com",
	}
	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			err := ValidateEmail(email)
			assert.Error(t, err)
		})
	}
}

// --- URL validation tests ---

func TestValidateURL_Valid(t *testing.T) {
	validURLs := []string{
		"https://github.com/user",
		"http://example.com",
		"https://linkedin.com/in/name",
	}
	for _, u := range validURLs {
		t.Run(u, func(t *testing.T) {
			err := ValidateURL(u)
			assert.NoError(t, err)
		})
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	invalidURLs := []string{
		"",
		"not a url",
		"ftp://example.com",
		"://missing-scheme",
	}
	for _, u := range invalidURLs {
		t.Run(u, func(t *testing.T) {
			err := ValidateURL(u)
			assert.Error(t, err)
		})
	}
}

// --- Competence level validation tests ---

func TestValidateCompetenceLevel_Valid(t *testing.T) {
	for level := 1; level <= 10; level++ {
		assert.True(t, ValidCompetenceLevel(level), "level %d should be valid", level)
	}
}

func TestValidateCompetenceLevel_Invalid(t *testing.T) {
	assert.False(t, ValidCompetenceLevel(0))
	assert.False(t, ValidCompetenceLevel(11))
	assert.False(t, ValidCompetenceLevel(-1))
}

// --- Status validation tests ---

func TestValidateStatus_AllValid(t *testing.T) {
	for _, status := range AllStatuses {
		assert.True(t, ValidStatus(status), "status %q should be valid", status)
	}
}

func TestValidateStatus_Invalid(t *testing.T) {
	assert.False(t, ValidStatus(""))
	assert.False(t, ValidStatus("NotAStatus"))
}

// --- Fit indicator validation tests ---

func TestValidateFitIndicator_AllValid(t *testing.T) {
	for _, fit := range AllFitIndicators {
		assert.True(t, ValidFitIndicator(fit), "fit %q should be valid", fit)
	}
}

func TestValidateFitIndicator_EmptyIsValid(t *testing.T) {
	assert.True(t, ValidFitIndicator(""), "empty fit indicator is valid (not assessed)")
}

func TestValidateFitIndicator_Invalid(t *testing.T) {
	assert.False(t, ValidFitIndicator("Bad Fit"))
}

// --- Granularity validation tests ---

func TestValidateGranularity_AllValid(t *testing.T) {
	for _, g := range AllGranularities {
		assert.True(t, ValidGranularity(g), "granularity %q should be valid", g)
	}
}

func TestValidateGranularity_Invalid(t *testing.T) {
	assert.False(t, ValidGranularity(""))
	assert.False(t, ValidGranularity("century"))
}

// --- WorkHistory input validation tests ---

func TestValidateWorkHistoryInput_Valid(t *testing.T) {
	input := WorkHistoryInput{
		EmployerName:         "Acme Corp",
		JobTitle:             "Developer",
		StartDate:            "2023-01",
		DateGranularityStart: "month",
	}
	err := ValidateWorkHistoryInput(input)
	assert.NoError(t, err)
}

func TestValidateWorkHistoryInput_MissingEmployer(t *testing.T) {
	input := WorkHistoryInput{
		EmployerName:         "",
		JobTitle:             "Developer",
		StartDate:            "2023",
		DateGranularityStart: "year",
	}
	err := ValidateWorkHistoryInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "employer")
}

func TestValidateWorkHistoryInput_EndBeforeStart(t *testing.T) {
	input := WorkHistoryInput{
		EmployerName:         "Acme",
		JobTitle:             "Dev",
		StartDate:            "2024",
		EndDate:              "2023",
		DateGranularityStart: "year",
		DateGranularityEnd:   "year",
	}
	err := ValidateWorkHistoryInput(input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end date")
}
