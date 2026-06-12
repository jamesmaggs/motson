package fixtures

import "sort"

// GroupStanding is a team's record within its group, derived from
// finished group-stage matches only (spec value GroupStanding).
type GroupStanding struct {
	Team         string
	Played       int
	Won          int
	Drawn        int
	Lost         int
	GoalsFor     int
	GoalsAgainst int
}

func (s GroupStanding) GoalDifference() int { return s.GoalsFor - s.GoalsAgainst }
func (s GroupStanding) Points() int         { return s.Won*3 + s.Drawn }

// Standings computes the table for the given matches. Every named
// team appears (zeroed until it has played); only finished matches
// contribute. Ranking follows the StandingsOrder guarantee: points,
// goal difference, goals scored, team name — a deliberate
// simplification of FIFA's head-to-head tiebreakers.
func Standings(matches []Match) []GroupStanding {
	table := map[string]*GroupStanding{}
	team := func(name string) *GroupStanding {
		if name == "" {
			return nil
		}
		if _, ok := table[name]; !ok {
			table[name] = &GroupStanding{Team: name}
		}
		return table[name]
	}

	for _, m := range matches {
		home, away := team(m.HomeTeam), team(m.AwayTeam)
		if m.Status != StatusFinished || m.HomeScore == nil || m.AwayScore == nil ||
			home == nil || away == nil {
			continue
		}
		hg, ag := *m.HomeScore, *m.AwayScore
		home.Played, away.Played = home.Played+1, away.Played+1
		home.GoalsFor += hg
		home.GoalsAgainst += ag
		away.GoalsFor += ag
		away.GoalsAgainst += hg
		switch {
		case hg > ag:
			home.Won++
			away.Lost++
		case hg < ag:
			away.Won++
			home.Lost++
		default:
			home.Drawn++
			away.Drawn++
		}
	}

	standings := make([]GroupStanding, 0, len(table))
	for _, s := range table {
		standings = append(standings, *s)
	}
	sort.Slice(standings, func(i, j int) bool {
		a, b := standings[i], standings[j]
		if a.Points() != b.Points() {
			return a.Points() > b.Points()
		}
		if a.GoalDifference() != b.GoalDifference() {
			return a.GoalDifference() > b.GoalDifference()
		}
		if a.GoalsFor != b.GoalsFor {
			return a.GoalsFor > b.GoalsFor
		}
		return a.Team < b.Team
	})
	return standings
}
