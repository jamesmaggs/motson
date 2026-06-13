---
target: index page
total_score: 35
p0_count: 0
p1_count: 0
timestamp: 2026-06-13T10-39-42Z
slug: internal-web-templates-index-html-tmpl
---
# Critique — Motson index page (after impeccable passes 1–3)

## Design Health Score

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 4 | LIVE badge (text) + last-updated time. |
| 2 | Match System / Real World | 4 | Flags, group/stage, local kickoff, amber score. |
| 3 | User Control and Freedom | 3 | Nav + calendar + collapsible menu; read-only. |
| 4 | Consistency and Standards | 4 | CSS token palette; consistent components; Inter (accepted). |
| 5 | Error Prevention | 3 | n/a — read-only. |
| 6 | Recognition Rather Than Recall | 4 | Everything visible: flags, group grid, team list, skip link. |
| 7 | Flexibility and Efficiency | 3 | Sidebar jump to group/team; no filter (P3). |
| 8 | Aesthetic and Minimalist Design | 4 | Dead space removed; compact, restrained, on-brand. |
| 9 | Error Recovery | 3 | n/a. |
| 10 | Help and Documentation | 3 | Self-explanatory glance page. |
| **Total** | | **35/40** | **Good (top band)** |

## Anti-Patterns Verdict
**LLM**: Distinctive and intentional — mascot, floodlit palette, Bebas title, gradient header border, broadcast LIVE badge. No glow, no gradient text, no glassmorphism, no side stripes.
**Deterministic scan**: 1 warning — `overused-font` (Inter). Accepted: explicit user choice and permitted for the product register; character is carried by Bebas + theme.

## What's Working
- Scores-first cards: amber score prominent, balanced flags, no dead space.
- Accessible: visible focus, skip-to-fixtures, per-card aria-labels, AA contrast throughout, reduced-motion guard.
- Coherent token-based palette; consistent interactive states with smooth 150ms transitions.

## Resolved since first critique (31 → 35)
- Card dead space (grid-auto-rows:1fr) removed.
- In-progress now a textual LIVE badge, not a colour-only glow.
- Muted text contrast 3.92:1 → 5.1:1 (AA).
- Focus indicators, skip link, card aria-labels, mascot dimensions added.
- Palette extracted to CSS custom properties.

## Remaining (accepted / P3)
- Inter font (accepted — user choice + register-permitted).
- No date-grouping / filter across 104 matches (P3 — sidebar covers group/team jumps; chronological list with LIVE highlighting serves the glance use-case).
