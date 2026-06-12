# 0011. Serve Motson at motson.jamesmaggs.com

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

Motson needs a stable public address for two long-lived consumers: the
calendar feed URL that subscribers add to Apple Calendar (effectively
immutable once shared — changing it breaks every subscription) and the
fixtures page. Railway provides a default `*.up.railway.app` domain and
supports custom domains.

Drivers:

- Feed URLs are subscribed to, so the address must not change
  mid-tournament
- The owner already holds the jamesmaggs.com domain
- A memorable address is easier to share with friends

## Considered options

- **Railway default domain** (`motson.up.railway.app`) — zero setup,
  but tied to the platform and less memorable
- **Custom subdomain** (`motson.jamesmaggs.com`) — one CNAME record and
  a Railway custom-domain entry; platform-independent address

## Decision

Serve Motson at `motson.jamesmaggs.com`: a CNAME from the subdomain to
the Railway service, registered as a custom domain in Railway (which
provisions TLS automatically).

## Consequences

- Calendar subscriptions and shared links survive a platform move; only
  DNS would change
- One manual setup step alongside the uptime monitor (ADR 0008): the
  CNAME record and the Railway custom-domain entry
- Absolute URLs in the feed or page (if any) must use the custom
  domain, sourced from config rather than hardcoded
