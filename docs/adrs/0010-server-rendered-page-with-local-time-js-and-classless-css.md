# 0010. Server-rendered page with local-time JS and classless CSS

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

The fixtures page is one server-rendered view of the match list
(`FixturesPage` surface). World Cup 2026 spans four-plus North American
timezones while visitors may be anywhere, so the rendered kickoff time
needs a policy. The page also needs styling without contradicting the
lean stack (ADR 0005).

Drivers:

- Visitors should read kickoff times without timezone arithmetic
- The calendar feed already localises (UTC events render in the
  subscriber's timezone); the page should feel consistent with it
- No build step, minimal assets

## Considered options

- **Times**: visitor-local via tiny vanilla JS vs fixed timezone (no JS)
  vs venue-local
- **Styling**: hand-written CSS vs classless CSS (e.g. Pico.css) vs
  Tailwind

## Decision

The page is rendered with `html/template`, emitting UTC timestamps; a
few lines of vanilla inline JS convert them to the browser's local
time. Styling comes from a classless CSS file (Pico.css) served as a
static asset, over semantic HTML.

## Consequences

- Page and calendar agree: everyone sees their own wall-clock time
- With JS disabled the page degrades to UTC times (rendered as the
  fallback text), still correct if less convenient
- No build step; total front-end assets are one CSS file and a few
  lines of script
- Pico.css is vendored (committed, not CDN-linked) so the page has no
  external runtime dependencies
