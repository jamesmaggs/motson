// Package fixtures holds Motson's domain model: World Cup matches
// mirrored from an external provider, as specified in
// docs/allium/motson.allium.
package fixtures

import "time"

// Config defaults from the spec's config block.
const (
	SyncInterval          = time.Hour
	GroupMatchDuration    = 2 * time.Hour
	KnockoutMatchDuration = 2*time.Hour + 45*time.Minute
	StalenessThreshold    = 3 * time.Hour
)

type Stage string

const (
	StageGroup        Stage = "group"
	StageRoundOf32    Stage = "round_of_32"
	StageRoundOf16    Stage = "round_of_16"
	StageQuarterFinal Stage = "quarter_final"
	StageSemiFinal    Stage = "semi_final"
	StageThirdPlace   Stage = "third_place"
	StageFinal        Stage = "final"
)

// Status is the match status the provider last reported. The lifecycle
// is owned by the provider; Motson mirrors it without enforcing
// transitions.
type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusInPlay    Status = "in_play"
	StatusFinished  Status = "finished"
	StatusPostponed Status = "postponed"
	StatusCancelled Status = "cancelled"
)

type Match struct {
	ProviderMatchID string
	HomeTeam        string
	AwayTeam        string
	KickoffAt       time.Time
	Venue           string
	Stage           Stage
	GroupName       string // present only when Stage is StageGroup
	Status          Status
	// Scores are present only when Status is StatusFinished.
	HomeScore *int
	AwayScore *int
	// Penalties are present only for finished matches decided by a shootout.
	HomePenalties *int
	AwayPenalties *int
}

// EndsAt is the calendar event end time: kickoff plus the configured
// duration for the match's stage.
func (m Match) EndsAt() time.Time {
	if m.Stage == StageGroup {
		return m.KickoffAt.Add(GroupMatchDuration)
	}
	return m.KickoffAt.Add(KnockoutMatchDuration)
}
