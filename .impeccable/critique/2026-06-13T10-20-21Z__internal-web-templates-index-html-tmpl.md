---
target: index page
total_score: 31
p0_count: 0
p1_count: 2
timestamp: 2026-06-13T10-20-21Z
slug: internal-web-templates-index-html-tmpl
---
# Critique — Motson index page

## Design Health Score

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 3 | Last-updated time shown; in-progress signalled by green glow only (no text). |
| 2 | Match System / Real World | 4 | Football language, flags, group/stage, local kickoff times. Strong. |
| 3 | User Control and Freedom | 3 | Nav + calendar + collapsible menu; read-only page, no traps. |
| 4 | Consistency and Standards | 3 | Consistent card treatment & links; generic tells (Inter, glow). |
| 5 | Error Prevention | 3 | n/a — read-only; staleness handled by /healthz. |
| 6 | Recognition Rather Than Recall | 4 | Everything visible/labelled: flags, group grid, team list. |
| 7 | Flexibility and Efficiency | 3 | Calendar subscribe + group/team nav; no jump-to-today / filter. |
| 8 | Aesthetic and Minimalist Design | 2 | Dead space: every card stretched to the tallest (grid-auto-rows:1fr); unplayed cards ~50% empty. |
| 9 | Error Recovery | 3 | n/a on this page. |
| 10 | Help and Documentation | 3 | Self-explanatory glance page; no help needed. |
| **Total** | | **31/40** | **Good** |

## Anti-Patterns Verdict

**LLM**: Not egregiously AI — genuine character (Motson mascot, floodlit palette, Bebas title, fading header border). Two tells remain: Inter (overused) and the green box-shadow glow on the dark theme. The dominant non-slop problem is the empty cards.

**Deterministic scan** (detect.mjs on rendered HTML): 2 warnings — `overused-font` (Inter, line 15) and `dark-glow` (green glow rgb(47,191,113), line 54).

## What's Working
- Strong real-world match: flags, group pills, kickoff in local time, amber score.
- Genuine, restrained character: mascot, floodlit night palette, Bebas display title, gradient header border.
- Solid navigation: Add-to-Calendar, group letter grid, team list; collapsible + full-width on mobile.

## Priority Issues
- **[P1] Dead space / low density**: `grid-auto-rows: 1fr` stretches every card to the tallest in the whole grid, so unplayed "vs" cards are ~half empty. Fix: let rows size to content (keep within-row equal height), tighten card padding. → layout
- **[P1] In-progress is colour-only (glow)**: fails "colour not the sole carrier" (a11y) and the glow is a detector tell. Add a textual LIVE indicator and drop the box-shadow glow. → colorize/clarify
- **[P2] Muted text contrast**: `.vs`/`.state` (#79838e) is 3.92:1 on the card — below WCAG AA (4.5). Bump toward the ink end. → audit
- **[P3] Inter overused**: product register permits Inter; accepted. Character carried by Bebas + theme.
- **[P3] No jump-to-today / filter** across 104 matches; sidebar covers groups/teams, so low impact.

## Persona Red Flags
- **Casey (distracted mobile)**: dead-space cards mean lots of scrolling to pass empty fixtures; a denser layout gets more matches per screen. Buttons are thumb-reachable; state persists (server-rendered).
- **Sam (accessibility)**: in-progress conveyed by colour alone (glow); muted `vs`/state text below AA contrast. Both fixable.

## Questions to Consider
- Could unplayed and played matches share one compact row height instead of stretching to the tallest?
- Should "live" read as a broadcast LIVE badge rather than a glow?
