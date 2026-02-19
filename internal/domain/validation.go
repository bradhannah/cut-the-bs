package domain

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateRequired checks that a field value is non-empty after
// trimming whitespace. Returns an error naming the field if empty.
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateEmail checks that the given string looks like a valid
// email address. This is a simple structural check (contains @,
// has local and domain parts), not a full RFC 5322 parser.
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid email format: %s", email)
	}
	// Domain must contain at least one dot and not start with one.
	if !strings.Contains(parts[1], ".") || strings.HasPrefix(parts[1], ".") {
		return fmt.Errorf("invalid email domain: %s", email)
	}
	return nil
}

// ValidateURL checks that the given string is a valid HTTP or HTTPS
// URL.
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %s", rawURL)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme: %s", rawURL)
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host: %s", rawURL)
	}
	return nil
}

// ValidateDateRange validates that end date is not before start date,
// compared at the coarsest granularity of the two dates. An empty
// end date means "present" and is always valid.
func ValidateDateRange(start, startGran, end, endGran string) error {
	if end == "" {
		return nil // "present" — always valid
	}

	// Normalize both dates to the coarsest granularity for comparison.
	startNorm := normalizeDateToCoarsest(start, startGran, endGran)
	endNorm := normalizeDateToCoarsest(end, endGran, startGran)

	if endNorm < startNorm {
		return fmt.Errorf("end date (%s) must not be before start date (%s)", end, start)
	}

	return nil
}

// normalizeDateToCoarsest truncates a date string to the coarsest
// of its own granularity and the other granularity, for comparison.
func normalizeDateToCoarsest(date, ownGran, otherGran string) string {
	coarsest := coarserGranularity(ownGran, otherGran)
	switch coarsest {
	case GranularityYear:
		// Take just the first 4 chars (the year).
		if len(date) >= 4 {
			return date[:4]
		}
		return date
	case GranularityMonth:
		// Take up to 7 chars (YYYY-MM).
		if len(date) >= 7 {
			return date[:7]
		}
		return date
	default:
		// Day — use full date.
		return date
	}
}

// coarserGranularity returns the coarser of two granularity values.
// Order from coarsest to finest: year > month > day.
func coarserGranularity(a, b string) string {
	rank := map[string]int{
		GranularityYear:  0,
		GranularityMonth: 1,
		GranularityDay:   2,
	}
	ra, ok := rank[a]
	if !ok {
		ra = 2 // default to finest
	}
	rb, ok := rank[b]
	if !ok {
		rb = 2
	}
	if ra <= rb {
		return a
	}
	return b
}

// ValidateWorkHistoryInput validates all fields of a WorkHistoryInput.
func ValidateWorkHistoryInput(input WorkHistoryInput) error {
	if err := ValidateRequired(input.EmployerName, "employer name"); err != nil {
		return err
	}
	if err := ValidateRequired(input.JobTitle, "job title"); err != nil {
		return err
	}
	if err := ValidateRequired(input.StartDate, "start date"); err != nil {
		return err
	}
	if err := ValidateRequired(input.DateGranularityStart, "start date granularity"); err != nil {
		return err
	}
	if !ValidGranularity(input.DateGranularityStart) {
		return fmt.Errorf("invalid start date granularity: %s", input.DateGranularityStart)
	}
	if input.EndDate != "" {
		if input.DateGranularityEnd == "" {
			return fmt.Errorf("end date granularity is required when end date is provided")
		}
		if !ValidGranularity(input.DateGranularityEnd) {
			return fmt.Errorf("invalid end date granularity: %s", input.DateGranularityEnd)
		}
		if err := ValidateDateRange(input.StartDate, input.DateGranularityStart, input.EndDate, input.DateGranularityEnd); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSkillInput validates all fields of a SkillInput.
func ValidateSkillInput(input SkillInput) error {
	if err := ValidateRequired(input.Name, "skill name"); err != nil {
		return err
	}
	if input.CategoryID <= 0 {
		return fmt.Errorf("category is required")
	}
	if !ValidCompetenceLevel(input.CompetenceLevel) {
		return fmt.Errorf("competence level must be between %d and %d", MinCompetenceLevel, MaxCompetenceLevel)
	}
	return nil
}

// ValidateAcademicInput validates all fields of an AcademicInput.
func ValidateAcademicInput(input AcademicInput) error {
	if err := ValidateRequired(input.Institution, "institution"); err != nil {
		return err
	}
	if err := ValidateRequired(input.CredentialType, "credential type"); err != nil {
		return err
	}
	if err := ValidateRequired(input.FieldOfStudy, "field of study"); err != nil {
		return err
	}
	if err := ValidateRequired(input.CompletionDate, "completion date"); err != nil {
		return err
	}
	if input.DateGranularity != "" && !ValidGranularity(input.DateGranularity) {
		return fmt.Errorf("invalid date granularity: %s", input.DateGranularity)
	}
	return nil
}

// ValidateCertificationInput validates all fields of a
// CertificationInput.
func ValidateCertificationInput(input CertificationInput) error {
	if err := ValidateRequired(input.Name, "certification name"); err != nil {
		return err
	}
	if err := ValidateRequired(input.IssuingBody, "issuing body"); err != nil {
		return err
	}
	if err := ValidateRequired(input.DateEarned, "date earned"); err != nil {
		return err
	}
	return nil
}

// ValidateSummaryInput validates all fields of a SummaryInput.
func ValidateSummaryInput(input SummaryInput) error {
	if err := ValidateRequired(input.Label, "label"); err != nil {
		return err
	}
	if err := ValidateRequired(input.BodyText, "body text"); err != nil {
		return err
	}
	return nil
}

// ValidateApplicationInput validates all fields of an
// ApplicationInput.
func ValidateApplicationInput(input ApplicationInput) error {
	if err := ValidateRequired(input.CompanyName, "company name"); err != nil {
		return err
	}
	if err := ValidateRequired(input.PositionTitle, "position title"); err != nil {
		return err
	}
	if err := ValidateRequired(input.DateApplied, "date applied"); err != nil {
		return err
	}
	if !ValidFitIndicator(input.FitIndicator) {
		return fmt.Errorf("invalid fit indicator: %s", input.FitIndicator)
	}
	return nil
}

// ValidateCoverLetterInput validates all fields of a
// CoverLetterInput.
func ValidateCoverLetterInput(input CoverLetterInput) error {
	if err := ValidateRequired(input.Title, "title"); err != nil {
		return err
	}
	if err := ValidateRequired(input.BodyText, "body text"); err != nil {
		return err
	}
	return nil
}

// ValidateProfileLinkInput validates all fields of a
// ProfileLinkInput.
func ValidateProfileLinkInput(input ProfileLinkInput) error {
	if err := ValidateRequired(input.Label, "label"); err != nil {
		return err
	}
	if err := ValidateURL(input.URL); err != nil {
		return err
	}
	return nil
}
