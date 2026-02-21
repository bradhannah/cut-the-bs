package domain

// Application status values (FR-023). Any transition between statuses
// is permitted.
const (
	StatusApplied             = "Applied"
	StatusAcknowledged        = "Acknowledged"
	StatusScreening           = "Screening"
	StatusPhoneScreen         = "Phone Screen"
	StatusInterviewScheduled  = "Interview Scheduled"
	StatusInterviewCompleted  = "Interview Completed"
	StatusTechnicalAssessment = "Technical Assessment"
	StatusFinalRound          = "Final Round"
	StatusOfferReceived       = "Offer Received"
	StatusOfferAccepted       = "Offer Accepted"
	StatusOfferDeclined       = "Offer Declined"
	StatusEmployerRejected    = "Employer Rejected"
	StatusUserWithdrawn       = "User Withdrawn"
	StatusGhosted             = "Ghosted"
	StatusOnHold              = "On Hold"
)

// AllStatuses is the complete list of valid application status values.
var AllStatuses = []string{
	StatusApplied,
	StatusAcknowledged,
	StatusScreening,
	StatusPhoneScreen,
	StatusInterviewScheduled,
	StatusInterviewCompleted,
	StatusTechnicalAssessment,
	StatusFinalRound,
	StatusOfferReceived,
	StatusOfferAccepted,
	StatusOfferDeclined,
	StatusEmployerRejected,
	StatusUserWithdrawn,
	StatusGhosted,
	StatusOnHold,
}

// ValidStatus returns true if the given status string is in the
// allowed set.
func ValidStatus(status string) bool {
	for _, s := range AllStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// Fit indicator values (FR-032).
const (
	FitUnlikely = "Unlikely"
	FitStretch  = "Stretch Fit"
	FitPossible = "Possible Fit"
	FitStrong   = "Strong Fit"
	FitPerfect  = "Perfect Fit"
)

// AllFitIndicators is the complete list of valid fit indicator values.
var AllFitIndicators = []string{
	FitUnlikely,
	FitStretch,
	FitPossible,
	FitStrong,
	FitPerfect,
}

// ValidFitIndicator returns true if the given fit indicator string
// is in the allowed set. An empty string is valid (fit not assessed).
func ValidFitIndicator(fit string) bool {
	if fit == "" {
		return true
	}
	for _, f := range AllFitIndicators {
		if f == fit {
			return true
		}
	}
	return false
}

// Date granularity values for work history and academic dates.
const (
	GranularityYear  = "year"
	GranularityMonth = "month"
	GranularityDay   = "day"
)

// AllGranularities is the complete list of valid date granularity
// values.
var AllGranularities = []string{
	GranularityYear,
	GranularityMonth,
	GranularityDay,
}

// ValidGranularity returns true if the given granularity string is
// in the allowed set.
func ValidGranularity(g string) bool {
	for _, v := range AllGranularities {
		if v == g {
			return true
		}
	}
	return false
}
