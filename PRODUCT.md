# Product

## Register

product

## Users

Football fans following the 2026 World Cup who want to glance at fixtures and
scores without friction. They arrive on a phone between matches or at a desk
with the page open in a tab, scan for "what's on / what's the score", maybe
tap into a group table or a team, and leave. Many subscribe the calendar feed
once and rarely return to the page itself. The job is fast orientation, not
deep exploration.

## Product Purpose

Motson mirrors World Cup 2026 fixtures and scores (synced every 10 minutes from
an abstract provider) and presents them two ways: an Apple-Calendar-compatible
feed and this web page. The page exists so a fan can answer "is there a game
on, and what happened" in a couple of seconds, then optionally drill into a
group standings table or a team's fixtures. Success is a page someone can read
at a glance and trust to be current.

## Brand Personality

Broadcast nostalgia: a warm nod to classic football commentary — the John
Motson cutout, the floodlit night-match palette, the sense of a calm voice
reading the scores. Characterful but understated; never kitsch, never loud.
Three words: **assured, warm, unfussy.** The page should feel like it knows
the football, not like it's selling you something.

## Anti-references

- The busy, ad-heavy, odds-pushing mainstream football portals (betting
  banners, autoplay video, interstitials). No clutter, no monetisation theatre.
- Generic SaaS-dashboard chrome (hero metric tiles, gradient accents, endless
  identical card grids).
- Loud TV-graphics maximalism — big shouty type and dramatic accents
  everywhere. Motson is the quiet, knowledgeable commentator, not the
  pre-match hype reel.

## Design Principles

- **Scores first.** Every design choice serves a fan glancing for a result or
  a kickoff time. If an element doesn't help that, it earns its place or goes.
- **Trustworthy and current.** The page must read as live and accurate —
  status and freshness are legible, never ambiguous.
- **Quiet character.** Personality comes from restraint plus a few deliberate
  touches (the mascot, the floodlit palette, the amber score), not from
  decoration applied everywhere.
- **Lightweight by construction.** Server-rendered Go with minimal vanilla JS;
  no heavy frameworks or animation libraries. Fast on a phone on a train.

## Accessibility & Inclusion

- WCAG AA: body and score text meet AA contrast on the dark theme; interactive
  controls have clear affordance and focus states.
- Respect `prefers-reduced-motion` for any animation (the menu open/close and
  any future motion need a reduced-motion path).
- Subtle motion only — nothing that competes with reading the scores.
- Colour is never the sole carrier of meaning (e.g. in-progress matches are
  signalled by more than the green glow).
