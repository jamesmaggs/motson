// Package fixtures holds Motson's domain model: World Cup matches
// mirrored from an external provider, as specified in
// docs/allium/motson.allium.
package fixtures

import (
	"fmt"
	"time"
)

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

var stageLabels = map[Stage]string{
	StageGroup:        "Group stage",
	StageRoundOf32:    "Round of 32",
	StageRoundOf16:    "Round of 16",
	StageQuarterFinal: "Quarter-final",
	StageSemiFinal:    "Semi-final",
	StageThirdPlace:   "Third place",
	StageFinal:        "Final",
}

// Label is the stage's human-readable name.
func (s Stage) Label() string { return stageLabels[s] }

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

// Validate enforces the spec's NonNegativeScores invariant: a finished
// match cannot carry a negative score.
func (m Match) Validate() error {
	for _, score := range []*int{m.HomeScore, m.AwayScore} {
		if score != nil && *score < 0 {
			return fmt.Errorf("match %s: negative score violates NonNegativeScores", m.ProviderMatchID)
		}
	}
	return nil
}

// EndsAt is the calendar event end time: kickoff plus the configured
// duration for the match's stage.
func (m Match) EndsAt() time.Time {
	if m.Stage == StageGroup {
		return m.KickoffAt.Add(GroupMatchDuration)
	}
	return m.KickoffAt.Add(KnockoutMatchDuration)
}
