package domain

// CompetenceLevel represents one level on the fixed 10-point
// competence scale. Each level has a numeric value, a short label,
// and descriptive criteria to help users self-assess.
type CompetenceLevel struct {
	Level       int    `json:"level"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// CompetenceLevels is the fixed 10-level competence scale, ordered
// from highest (10) to lowest (1). Per FR-030, each level includes
// descriptive criteria.
var CompetenceLevels = []CompetenceLevel{
	{Level: 10, Label: "Visionary", Description: "Could define the industry direction for this technology; recognized thought leader"},
	{Level: 9, Label: "Authority", Description: "Could teach a class on this topic; deep expertise across edge cases and internals"},
	{Level: 8, Label: "Expert", Description: "Considered an expert among peers; go-to person for complex problems"},
	{Level: 7, Label: "Advanced", Description: "Highly proficient; can architect solutions and mentor others"},
	{Level: 6, Label: "Proficient", Description: "Productive independently; solid understanding of best practices"},
	{Level: 5, Label: "Competent", Description: "Used effectively in production environments; comfortable with common patterns"},
	{Level: 4, Label: "Developing", Description: "Can complete tasks with occasional guidance; growing confidence"},
	{Level: 3, Label: "Familiar", Description: "Used in projects but requires regular reference to documentation"},
	{Level: 2, Label: "Beginner", Description: "Used in a learning project or tutorial; basic understanding only"},
	{Level: 1, Label: "Awareness", Description: "Awareness only; knows it exists but has never used it"},
}

// MinCompetenceLevel is the lowest valid competence level value.
const MinCompetenceLevel = 1

// MaxCompetenceLevel is the highest valid competence level value.
const MaxCompetenceLevel = 10

// ValidCompetenceLevel returns true if the given level is within the
// valid range [1, 10].
func ValidCompetenceLevel(level int) bool {
	return level >= MinCompetenceLevel && level <= MaxCompetenceLevel
}

// GetCompetenceLevel returns the CompetenceLevel for the given
// numeric value, or nil if the value is out of range.
func GetCompetenceLevel(level int) *CompetenceLevel {
	for i := range CompetenceLevels {
		if CompetenceLevels[i].Level == level {
			return &CompetenceLevels[i]
		}
	}
	return nil
}
