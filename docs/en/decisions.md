**🌐 English** · [Español](../es/decisions.md)

# Design decisions

Each entry records the **context**, the **decision**, the **alternatives** that were weighed and the **consequences** we accept. When a decision changes, add a new entry instead of rewriting history.

| # | Decision | Status |
|---|---|---|
| D1 | Modular monolith with Clean Architecture | Accepted |
| D2 | Go + Gin for the backend | Accepted |
| D3 | PostgreSQL as the source of truth, Redis for ephemeral state | Accepted |
| D4 | Versioned SQL migrations (golang-migrate), applied on startup | Accepted |
| D5 | Short access JWT + rotating refresh token in an httpOnly cookie | Accepted |
| D6 | Access token only in memory on the client | Accepted |
| D7 | WebSocket authenticated with a single-use ticket | Accepted |
| D8 | Redis pub/sub as the cross-instance event bus | Accepted |
| D9 | One rank-based permission matrix | Accepted |
| D10 | Direct messages are private rooms | Accepted |
| D11 | Cursor pagination | Accepted |
| D12 | Full-text search inside PostgreSQL | Accepted |
| D13 | Soft delete for messages, anonymisation for accounts | Accepted |
| D14 | Global roles granted only through a CLI | Accepted |
| D15 | Dependency-free i18n and CSS-variable theming | Accepted |
| D16 | Optimistic sending with explicit states | Accepted |
| D17 | Visual direction "Foco" | Accepted |
| D18 | People directory, discoverable rooms and personal invitations | Accepted |
| D19 | Container images built in CI, not published; cloud-neutral CI/CD | Accepted |
| D20 | Persistent sessions with a sliding 7-day window | Accepted |
| D21 | Password reset by emailed single-use link | Accepted |
| D22 | pnpm as the frontend package manager | Accepted |
| D23 | Typed tokens and per-request session validation | Accepted |
| D24 | Per-user presence keys and a bounded hub | Accepted |
| D25 | Security headers and CSP on the web app | Accepted |
| D26 | Accessibility patterns: live announcements, in-app dialogs, focus handling | Accepted |
| D27 | Server limits and ordered shutdown | Accepted |
| D28 | Supply-chain policy for the frontend install | Accepted |

---

## D1. Modular monolith with Clean Architecture

- **Context:** a small team, one deployable product, but a domain (auth, rooms, messages, moderation) that will keep growing.
- **Decision:** a single Go process with one package per module; each module has `domain`, `application`, `persistence` and `presentation` layers. Modules talk to each other through small interfaces (for example `WebsocketBroadcaster`, `RoomLister`), never through each other's packages.
- **Alternatives:** microservices (operational cost with no benefit at this size); a flat "handlers + models" layout (fast to start, but business rules leak into HTTP code and become untestable).
- **Consequences:** use cases are tested with fakes and no database. Interfaces cost a little boilerplate and wiring is manual in `main.go`. If a module ever needs to be extracted, its seams already exist.

## D2. Go + Gin

- **Context:** a service dominated by many concurrent, mostly idle connections.
- **Decision:** Go for cheap goroutines and a single static binary; Gin for routing and middleware.
- **Alternatives:** Node.js (shares language with the frontend, but CPU-bound work and memory per connection scale worse); a heavier framework (more features we do not need).
- **Consequences:** the frontend and backend do not share types, so API contracts are kept by hand (an OpenAPI contract is on the roadmap).

## D3. PostgreSQL as source of truth, Redis for ephemeral state

- **Decision:** everything that must survive a restart lives in PostgreSQL. Everything with a TTL or that is shared between instances lives in Redis: refresh-token ids, WebSocket tickets, presence, rate counters and the event bus.
- **Alternatives:** keeping presence in the process (breaks with more than one instance); storing sessions in PostgreSQL (extra write load for data that is meant to expire).
- **Consequences:** two stateful dependencies to operate. The rate limiter **fails open**: if Redis is down, requests are allowed and the problem is logged, because availability matters more than strictness there.

## D4. Versioned SQL migrations

- **Context:** the first version created tables by hand from a `schema.sql`.
- **Decision:** golang-migrate files (`NNNNNN_name.up.sql` / `.down.sql`) embedded in the binary and applied on startup. A legacy hand-built database is adopted at version 1.
- **Rules:** never edit an applied migration; add a new numbered pair.
- **Consequences:** a fresh clone, CI and production converge on the same schema. Starting a new instance can briefly run migrations, which is acceptable at this scale.

## D5. Short access JWT + rotating refresh token

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API
    participant R as Redis
    C->>A: POST /auth/login
    A->>R: store refresh jti
    A-->>C: access JWT (1 h) in body + refresh JWT in httpOnly cookie
    Note over C: access token expires
    C->>A: POST /auth/refresh-token (cookie)
    A->>R: consume old jti (single use)
    A->>R: store new jti
    A-->>C: new access JWT + new refresh cookie
```

- **Decision:** a 1-hour access token returned in the body, and a 7-day refresh token whose id (`jti`) is stored in Redis and can be used **once**. Using it issues a new pair.
- **Why rotation:** a stolen refresh token stops working as soon as the legitimate client refreshes, and a replayed one is detectable. `POST /auth/logout` revokes the current id.
- **Alternatives:** long-lived access tokens (a leak is valid for days); server-side sessions (simpler revocation, but needs sticky state on every request).

## D6. Access token only in memory on the client

- **Decision:** the browser keeps the access token in a JavaScript variable (`lib/tokenStorage.ts`), never in `localStorage`. The refresh token is an httpOnly, `SameSite=Strict` cookie scoped to `/api/v1/auth`. On reload, `AuthContext` restores the session by calling refresh.
- **Why:** any XSS can read `localStorage`; it cannot read an httpOnly cookie, and the in-memory token disappears with the page.
- **Consequences:** one extra request on page load, and the axios interceptor must transparently refresh on `401`.

## D7. WebSocket authenticated with a single-use ticket

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API
    participant R as Redis
    C->>A: POST /ws-ticket (Bearer JWT)
    A->>R: SET ticket (TTL 30 s)
    A-->>C: ticket
    C->>A: GET /ws?ticket=…  (Origin checked)
    A->>R: GETDEL ticket
    A-->>C: connection upgraded
```

- **Context:** browsers cannot set an `Authorization` header on a WebSocket, so the token would end up in the URL, where proxies and logs record it.
- **Decision:** exchange the JWT for a 30-second, single-use ticket. The server also checks the `Origin` header against `CORS_ALLOWED_ORIGINS`.
- **Consequences:** a leaked URL is useless within seconds and cannot be replayed.

## D8. Redis pub/sub as the event bus

- **Decision:** the hub publishes every event to one Redis channel; each instance delivers it to its own connected clients. Presence is a Redis sorted set shared by all instances.
- **Alternatives:** sticky sessions (uneven load, and a room's members may sit on different instances anyway); Kafka/NATS (more infrastructure than the problem needs).
- **Consequences:** delivery is at-most-once. Clients reconnect with backoff and reload recent history to recover what they missed.

## D9. One rank-based permission matrix

- **Context:** moderation, pinning, invitations and deletion each had their own role comparison, which is where authorization bugs hide.
- **Decision:** `internal/shared/permissions` is the only place that decides who may do what. Room roles (`member < moderator < admin < owner`) and global roles map to a numeric rank, and rules read "the actor must **strictly outrank** the target".
- **Rule for contributors:** new checks go through that package, never through ad-hoc string comparisons. The frontend mirrors the matrix in `lib/permissions.ts` only to decide what to *show*; the server always decides.

## D10. Direct messages are private rooms

- **Decision:** a one-to-one chat is a private room with `is_direct = true`, one per pair of users (`direct_rooms`, with `user_a < user_b` enforced by a `CHECK`).
- **Why:** history, unread counters, search, reactions, presence and WebSocket delivery work with **no new code**.
- **Consequences:** a few features branch on `is_direct` (no member list, no reporting).

## D11. Cursor pagination

- **Decision:** history is paged by cursor (`?limit=&before=`) over `(room_id, created_at DESC)`.
- **Alternatives:** `OFFSET` pagination skips or repeats rows when new messages arrive and gets slower with depth.

## D12. Full-text search inside PostgreSQL

- **Decision:** a GIN index on `to_tsvector('simple', content)` with accent-insensitive matching (`unaccent`), limited to rooms the user belongs to.
- **Alternatives:** Elasticsearch/Meilisearch give better ranking but add a service to run and keep in sync. Revisit if search becomes a core feature.

## D13. Soft delete for messages, anonymisation for accounts

- **Decision:** deleting a message sets `deleted_at`, so threads, reports and the audit trail stay coherent. Deleting an account passes owned rooms to another member (or removes them if the owner is alone), and keeps messages shown as "deleted user" (`author_id` becomes `NULL`).
- **Why:** it honours a person's right to erase their data without breaking other people's conversations. `GET /users/export` provides the matching data export.

## D14. Global roles only through a CLI

- **Decision:** `users.role` (`user | moderator | admin`) is never writable through the HTTP API. It is granted with `cmd/admin` (`make admin-grant`).
- **Why:** there is no endpoint to attack for privilege escalation; whoever can reach the database decides who is an admin.

## D15. Dependency-free i18n and CSS-variable theming

- **Decision:** typed dictionaries (`en.ts` is the source of truth for keys, `es.ts` must satisfy the same type) and a `useT()` hook; light and dark themes through CSS variables and semantic Tailwind tokens.
- **Why:** the app has two languages and a few hundred strings, which does not justify a routing-aware i18n framework. A missing translation is a compile error, and a test checks that placeholders match in both languages.
- **Consequences:** the language is chosen on the client, so server-rendered HTML is English and may flash briefly before hydration. If SEO-relevant public pages appear, revisit with locale-prefixed routes.

## D16. Optimistic sending with explicit states

- **Decision:** a message appears immediately as *sending*, and becomes *not sent* with **Retry** and **Discard** on failure.
- **Why:** unstable networks are normal on mobile; hiding failures loses messages silently.

## D17. Visual direction "Foco"

- **Context:** the first interface was a neutral placeholder. Three directions were prototyped with the same content (desktop chat, mobile and sign-in): **Foco** (dense, team tool), **Calidez** (rounded, bubbles, community) and **Redacción** (editorial, serif, monochrome).
- **Decision:** Foco. Bordered message cards with author and role, a left room list with `#` prefixes, a members side panel and a single indigo accent.
- **Why:** the product's centre of gravity is rooms, roles and moderation, which need information density and clear role labels more than personality. Cards show author, role and time compactly, and one accent keeps unread counts and selection unambiguous in both themes.
- **Alternatives:** Calidez has the friendliest mobile experience but spends width on bubbles; Redacción is the most distinctive but is harder to keep consistent as features grow.
- **Consequences:** all colour comes from semantic tokens (`primary`, `soft`, `success`...), so the direction can still be retuned in one file (`globals.css`).

## D18. People directory, discoverable rooms and personal invitations

- **Context:** until now you could only reach someone through a room you shared, and the only way into a private room was an anonymous code sent by hand through the API.
- **Decision:** three separate ideas. (1) *Presence is global:* an always-visible people panel lists everyone with online status, and you can write to anyone (a direct conversation needs no shared room). (2) *Rooms are discoverable:* every room is listed; public ones can be joined, private ones are visible but their messages are readable only by members. (3) *Personal invitations:* an owner or admin invites a specific person; it appears in that person's app and, only once accepted, makes them a member with history and write access.
- **Why:** being online, talking privately and belonging to a room are different questions. Mixing them (members only visible inside rooms) made the app feel empty. Keeping privacy at the room level protects private content while still advertising that the room exists.
- **Alternatives:** hiding private rooms entirely (nothing to ask an invitation for), or invite codes only (clumsy and unaddressed).
- **Consequences:** the directory exposes names and usernames to every signed-in user, which suits a team server; for a public deployment it would need a visibility setting. Invitations are polled every 15 s rather than pushed.

## D19. Container images built in CI, not published; cloud-neutral CI/CD

- **Context:** CI only checked the code. Nothing produced a runnable, versioned artifact, the frontend had no Dockerfile, the migrations had never been proven on an empty database, and dependencies were never scanned.
- **Decision:** CI also builds both images on every run to prove the Dockerfiles work, but does not publish them anywhere (an earlier version pushed multi-architecture images to GHCR from `main` and tags; removed on request); a separate weekly workflow scans dependencies; Dependabot keeps everything current. Deployment itself is deliberately not in the pipeline.
- **Why:** an image runs on any platform, so choosing AWS later (or something else) does not invalidate the work. Building images on pull requests catches a broken Dockerfile before merge. The first scan found real problems (critical Next.js advisories, an end-of-life Go line with 29 reachable vulnerabilities, an unused database driver with an unfixable one), which is the argument for running it continuously.
- **Alternatives:** deploy scripts for one provider (not portable), or publishing to a registry (dropped: not wanted for now; it can return later with an OIDC-authenticated login).
- **Consequences:** the web image is per environment because Next.js inlines `NEXT_PUBLIC_*` at build time. The Go toolchain moved from 1.24 (no longer patched) to 1.26, and the unused `pgx` driver was removed; the real driver is `lib/pq`. See [CI/CD](ci-cd.md).

## D20. Persistent sessions with a sliding 7-day window

- **Context:** people expect to come back to an open app after closing the tab, and to be asked to sign in again only after a long absence.
- **Decision:** the refresh token lives 7 days in a persistent httpOnly cookie, and **every renewal issues a new token with a fresh 7 days**. So a session ends after seven days *without use*, not seven days after sign-in. A used token stays valid for 30 seconds more. The browser signs the user out only when the server explicitly rejects the token (401/403), never because the network failed.
- **Why:** a fixed 7-day session would sign out someone who uses the app daily. The 30-second grace covers two tabs restoring at once and a response lost in transit, which would otherwise look like token theft and sign the user out. Treating a network failure as a rejection would end valid sessions every time a laptop wakes up offline.
- **Alternatives:** a longer fixed lifetime (signs out active users anyway, and widens the theft window), or never expiring (a lost device stays signed in forever).
- **Consequences:** a used token is replayable for 30 seconds, a deliberate trade for reliability. Activity means the app renewing its access token (about hourly while in use); a tab left open and untouched does not extend the session. Refresh tokens live in Redis, so losing Redis without persistence signs everyone out.

## D21. Password reset by emailed single-use link

- **Context:** forgetting the password left the account unreachable. The mailer already existed, but email *verification* is paused (H-008).
- **Decision:** `POST /auth/forgot-password` emails a link with a random token valid for 30 minutes; `POST /auth/reset-password` redeems it. The token is random bytes, not a JWT, and only its SHA-256 hash is stored in Redis. Redeeming it sets the password under the same policy as registration, burns the link, and rejects every refresh token issued before that moment.
- **Why:** the endpoint answers identically whether or not the address exists, otherwise it becomes a way to list who has an account. Storing a hash means a Redis leak yields no usable links. Ending all sessions matters because the usual reason to reset a password is a suspicion that someone else has it. A one-minute cooldown per address and a rate limit per client stop the form being used to flood someone's inbox.
- **Alternatives:** a signed JWT link (cannot be revoked or burned without storing something anyway), or security questions (weak and guessable).
- **Consequences:** it needs working SMTP settings; if sending fails the person still sees the neutral confirmation (the error is only logged). Already-issued access tokens stay valid until they expire, at most an hour. Link scanners that prefetch URLs are harmless because opening the page does not redeem the token; only submitting the form does.

## D22. pnpm as the frontend package manager

- **Context:** the frontend's package manager was a separate runtime used only to install and run scripts, while the app itself runs on Node (Next.js, the container, Vitest): one more thing to install locally, in CI and in the image.
- **Decision:** pnpm, pinned through `packageManager` in `package.json` (so Corepack, `pnpm/action-setup` and the Dockerfile all use the same version). Lockfile `pnpm-lock.yaml`; the install script of `unrs-resolver` is explicitly not allowed (`pnpm-workspace.yaml`), since it ships prebuilt binaries. CI, the security scan and the image build all install with `--frozen-lockfile` on Node 24 (`engines` requires 24.15 or newer); Dependabot uses the `npm` ecosystem, which reads pnpm lockfiles.
- **Why:** one runtime (Node) everywhere, strict `node_modules` that surfaces undeclared dependencies, a content-addressed store, and `pnpm audit` for the security scan.
- **Consequences:** the image needs only Node. Everyone needs pnpm locally (`corepack enable` on Node below 25, otherwise `npm i -g pnpm`).

## D23. Typed tokens and per-request session validation

- **Context:** the email-verification JWT shared the access-token secret and carried no marker, so the middleware accepted it as an access token; deactivated users and users whose sessions were revoked kept API access until their access token expired (up to an hour).
- **Decision:** every JWT carries a `typ` claim (`access`, `refresh`, `verify`) and each consumer requires its own type; signing is pinned to HS256. `JWTMiddleware` takes a `SessionGuard` that rejects tokens issued before the user's `sessions_revoked_at` (Redis) and tokens of deactivated users (Postgres, cached 15 s); it fails open and logs when a dependency is down. Deactivation now needs the password, revokes sessions and disconnects live sockets; password reset also disconnects sockets, and the WebSocket ticket redemption consults the same guard.
- **Why:** a token type must be checked by its consumer, not inferred from the absence of another claim. Stateless JWTs are fine until you need to cut access immediately, which a small per-request check buys at the price of one Redis read.
- **Consequences:** sessions open at deploy time get one 401 and refresh once. Usernames are limited to `[A-Za-z0-9_.-]` and email and username are unique case-insensitively (migration 13 refuses to run if two accounts collide, listing them). Login does a dummy bcrypt compare for unknown users. A ticket does not carry its issue time, so a ticket issued before a reset can still connect within its 30 s life.

## D24. Per-user presence keys and a bounded hub

- **Context:** presence lived in one sorted set read in full on every connect, disconnect and listing (about 4 GB of Redis traffic in a 10k-user reconnect storm); the hub scanned every room on each disconnect and held a global lock while delivering.
- **Decision:** one sorted set per user (`webchat:presence:u:<id>`, scored by expiry) updated by atomic Lua scripts, plus a secondary index `webchat:presence:online` (one member per online user, read with a LIMIT) for "who is online". The hub keeps per-client room sets and a user index, delivers from a snapshot outside the lock, marshals once, never blocks on a full broadcast channel (it drops and counts) and has `Shutdown` that closes sockets with code 1012.
- **Why:** every decision now touches one user's key, so cost does not grow with the number of connected users, and a slow consumer or a Redis stall cannot freeze request handlers.
- **Consequences:** a crashed instance's users stay online up to two minutes (TTL). `OnlineUsers` checks at most 1000 ids per call.

## D25. Security headers and CSP on the web app

- **Decision:** `next.config.mjs` sets `X-Frame-Options: DENY`, `nosniff`, `Referrer-Policy: no-referrer` (the reset token is in the URL), a restrictive `Permissions-Policy`, HSTS in production and a CSP whose `connect-src` is derived at build time from `NEXT_PUBLIC_API_URL` (HTTP and WebSocket origins).
- **Why `unsafe-inline`:** the theme and language init script must run before first paint, and a nonce would force dynamic rendering of every page and needs a proxy. `unsafe-eval` is added only in development.
- **Consequences:** the web image is built per environment (already true for the API URL); HSTS is honoured only over HTTPS; emoji-mart's styles are covered by `style-src 'unsafe-inline'`.

## D26. Accessibility patterns

- **Decision:** new messages from others are announced through one visually hidden `role="status"` element, coalesced per second, instead of `aria-live` on the whole list. Native `prompt`/`confirm` became an in-app `ConfirmDialog` (`alertdialog`, focus trap, Escape, restored focus). Dialogs use `useFocusTrap`; the emoji picker uses guard spans because it lives in a shadow root. The moderation panel is a complete ARIA tablist. The tab title is composed of a translated page title plus an unread prefix (`lib/documentTitle.ts`). Hover-only actions appear on devices without hover (`@media (hover:hover)`).
- **Contrast:** input borders, `success` text, the off-state of switches and the offline presence dot were raised to WCAG thresholds (new `presence-offline` token); a global `prefers-reduced-motion` rule disables non-essential motion.
- **Limit:** none of this has been verified with a screen reader or in several browsers yet.

## D27. Server limits and ordered shutdown

- **Decision:** the HTTP server has read-header, read, write and idle timeouts and a 64 KiB header cap; request bodies are limited to 1 MiB (413 when the declared length is over, 400 for a chunked body). Routes that mutate state have per-IP rate limits. `TRUSTED_PROXIES` (comma-separated IPs/CIDRs, empty by default) enables `X-Forwarded-For` behind a load balancer, otherwise all users share one bucket. Shutdown runs in order: signal, close sockets with 1012, drain HTTP, then close DB and Redis, with a 15 s cap and no `log.Fatal`. SMTP sends have dial and conversation deadlines.
- **Why:** without timeouts a slow client holds a connection forever; without trusted proxies, rate limiting behind a load balancer is global.
- **Consequences:** WebSocket connections are unaffected by the write timeout (the upgrade clears deadlines).

## D28. Supply-chain policy for the frontend install

- **Decision:** `frontend/pnpm-workspace.yaml` sets `minimumReleaseAge: 1440` (a version must be a day old), `trustPolicy: no-downgrade` (a release may not lose provenance compared with an earlier one), `blockExoticSubdeps` (git/tarball sources only for direct dependencies) and `strictDepBuilds` (an unreviewed build script fails the install; `allowBuilds` records the decisions). Three backport releases that trip the trust check are excluded by exact version in `trustPolicyExclude`.
- **Why:** pnpm re-checks every lockfile entry against these policies on each install, so they also guard `--frozen-lockfile` in CI and in the image, not only the moment someone adds a package.
- **Consequences:** a lockfile that contains a version younger than a day fails to install until it ages (or is listed in `minimumReleaseAgeExclude` for an urgent security fix). `tailwind-merge` stays on 2.x because it is the line for Tailwind 3; Dependabot will keep proposing 3.x until Tailwind 4 is adopted.
