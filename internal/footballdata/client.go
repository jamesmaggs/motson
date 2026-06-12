// Package footballdata adapts football-data.org's v4 API onto the
// FixtureSource contract (ADR 0003). It is the only code that knows
// the provider's shape; everything downstream speaks fixtures.Match.
package footballdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jazzatola/motson/internal/fixtures"
)

type Client struct {
	baseURL     string
	token       string
	competition string
	http        *http.Client
}

func New(baseURL, token, competition string) *Client {
	return &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		token:       token,
		competition: competition,
		http:        &http.Client{Timeout: 30 * time.Second},
	}
}

type matchesResponse struct {
	Matches []providerMatch `json:"matches"`
}

type providerMatch struct {
	ID       int64     `json:"id"`
	UTCDate  time.Time `json:"utcDate"`
	Status   string    `json:"status"`
	Stage    string    `json:"stage"`
	Group    *string   `json:"group"`
	Venue    string    `json:"venue"`
	HomeTeam team      `json:"homeTeam"`
	AwayTeam team      `json:"awayTeam"`
	Score    score     `json:"score"`
}

type team struct {
	Name string `json:"name"`
}

type score struct {
	FullTime  scorePair  `json:"fullTime"`
	Penalties *scorePair `json:"penalties"`
}

type scorePair struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

var statusMap = map[string]fixtures.Status{
	"SCHEDULED": fixtures.StatusScheduled,
	"TIMED":     fixtures.StatusScheduled,
	"IN_PLAY":   fixtures.StatusInPlay,
	"PAUSED":    fixtures.StatusInPlay,
	"FINISHED":  fixtures.StatusFinished,
	"AWARDED":   fixtures.StatusFinished,
	"POSTPONED": fixtures.StatusPostponed,
	"SUSPENDED": fixtures.StatusPostponed,
	"CANCELLED": fixtures.StatusCancelled,
}

var stageMap = map[string]fixtures.Stage{
	"GROUP_STAGE":    fixtures.StageGroup,
	"LAST_32":        fixtures.StageRoundOf32,
	"LAST_16":        fixtures.StageRoundOf16,
	"QUARTER_FINALS": fixtures.StageQuarterFinal,
	"SEMI_FINALS":    fixtures.StageSemiFinal,
	"THIRD_PLACE":    fixtures.StageThirdPlace,
	"FINAL":          fixtures.StageFinal,
}

// FetchFixtures implements the FixtureSource contract: one call
// returns the provider's complete tournament snapshot.
func (c *Client) FetchFixtures(ctx context.Context) ([]fixtures.Match, error) {
	url := fmt.Sprintf("%s/v4/competitions/%s/matches", c.baseURL, c.competition)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling football-data.org: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("football-data.org returned %s", resp.Status)
	}

	var body matchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding football-data.org response: %w", err)
	}

	matches := make([]fixtures.Match, 0, len(body.Matches))
	for _, pm := range body.Matches {
		m, err := mapMatch(pm)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func mapMatch(pm providerMatch) (fixtures.Match, error) {
	status, ok := statusMap[pm.Status]
	if !ok {
		return fixtures.Match{}, fmt.Errorf("match %d: unknown provider status %q", pm.ID, pm.Status)
	}
	stage, ok := stageMap[pm.Stage]
	if !ok {
		return fixtures.Match{}, fmt.Errorf("match %d: unknown provider stage %q", pm.ID, pm.Stage)
	}

	m := fixtures.Match{
		ProviderMatchID: strconv.FormatInt(pm.ID, 10),
		HomeTeam:        pm.HomeTeam.Name,
		AwayTeam:        pm.AwayTeam.Name,
		KickoffAt:       pm.UTCDate,
		Venue:           pm.Venue,
		Stage:           stage,
		GroupName:       groupName(pm.Group),
		Status:          status,
	}
	// Scores are present only on finished matches (spec field presence).
	if status == fixtures.StatusFinished {
		m.HomeScore, m.AwayScore = pm.Score.FullTime.Home, pm.Score.FullTime.Away
		if pm.Score.Penalties != nil {
			m.HomePenalties, m.AwayPenalties = pm.Score.Penalties.Home, pm.Score.Penalties.Away
		}
	}
	return m, nil
}

// groupName turns the provider's "GROUP_A" into "Group A".
func groupName(group *string) string {
	if group == nil {
		return ""
	}
	if letter, ok := strings.CutPrefix(*group, "GROUP_"); ok {
		return "Group " + letter
	}
	return *group
}
