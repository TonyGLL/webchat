**🌐 English** · [Español](../es/security.md)

# Security

Security is layered: no single control is trusted on its own.

```mermaid
flowchart LR
    R[Request] --> O[Origin / CORS]
    O --> RL[Rate limit<br/>per IP · Redis]
    RL --> J[JWT validation]
    J --> P[Permission matrix<br/>rank-based]
    P --> V[Input validation]
    V --> DB[(Data)]
    P -. every moderation action .-> AU[(audit_log)]
```

## 1. Authentication

| Item | Behaviour |
|---|---|
| Passwords | bcrypt hashes in their own table (`passwords`), never returned by the API |
| Password policy | 8–72 characters (72 is bcrypt's limit); rejects a short list of very common passwords and passwords containing the username or the email's local part |
| Access token | JWT, 1 hour, sent as `Authorization: Bearer …` |
| Refresh token | JWT, 7 days, **single use**; id stored in Redis; httpOnly, `SameSite=Strict` cookie scoped to `/api/v1/auth` |
| Logout | `POST /auth/logout` revokes the refresh id and clears the cookie |
| Secret | the server refuses to start without a `JWT_SECRET` of at least 32 characters |
| Login errors | the login form shows one generic message, so it does not reveal whether an email exists |

The reasoning behind each choice is in [Design decisions](decisions.md) (D5–D7).

## 2. WebSocket

- The JWT never appears in a URL: `POST /ws-ticket` returns a 30-second, single-use ticket stored in Redis.
- The `Origin` header must match `CORS_ALLOWED_ORIGINS`.
- A client can only announce `typing` in a room it is subscribed to, and no more often than a fixed interval.
- When a member is removed or banned, their connection is unsubscribed from the room (or disconnected on account deletion).

## 3. Authorization: the permission matrix

Everything lives in `internal/shared/permissions`. Room roles and global roles collapse into one **rank**:

```mermaid
flowchart LR
    N["0 · none"] --> M["1 · member"] --> MO["2 · moderator"] --> AD["3 · admin"] --> OW["4 · owner"] --> SU["5 · global admin"]
```

A global **moderator** acts as a moderator in every room; a global **admin** outranks every room owner.

| Action | Minimum rank |
|---|---|
| Read a room, send messages | 1 · member |
| Delete other people's messages, resolve reports | 2 · moderator |
| Kick or ban a user | 2 · moderator **and strictly above the target** |
| Pin messages, create invitations (codes or personal) | 3 · admin |
| Change roles | 3 · admin, only for members below them: admins can make moderators, only owners (or global admins) can make admins; the owner role is never assigned |
| Read the audit log | 3 · admin |

Because rules compare ranks, a moderator can never act on another moderator, an admin, or the owner.

## 4. Abuse protection

Rate limits are per client IP, kept in Redis (`/api/v1`):

| Route | Limit |
|---|---|
| `POST /auth/login` | 10 per minute |
| `POST /auth/register` | 20 per 10 minutes |
| `POST /auth/refresh-token` | 30 per minute |
| `POST /auth/send-verify-email/:user_id` | 5 per 10 minutes |
| `POST /ws-ticket` | 30 per minute |
| `POST /messages` | 120 per minute |
| `GET /messages/search` | 60 per minute |
| `GET /users/export`, `DELETE /users/me` | 5 per hour each |

When exceeded the API answers `429` with a `Retry-After` header. If Redis is unavailable the limiter lets traffic through and logs the failure. Behind a reverse proxy, configure trusted proxies so the client IP cannot be spoofed.

## 5. Moderation and audit

- Members report messages (one report per person, never your own). Moderators review the queue and can delete the message, mark the report resolved or ignore it.
- Bans are per room, temporary or permanent, and prevent re-joining by id or invitation.
- Every sensitive action is written to `audit_log` with actor, target, room and details. Only admins read it.

## 6. Privacy

- **Directory and discovery:** any signed-in user can list other active users (name, username, online status) and every room's name, topic and member count. Reading or writing in a room still needs membership, so a private room is visible but closed. Personal invitations can only be answered by the invited user (anyone else gets "not found"), and accepting re-checks bans.
- `GET /users/export` downloads a JSON file with the user's profile, rooms, messages and reactions.
- `DELETE /users/me` (confirmed with the password) deletes the account: see D13 in [Design decisions](decisions.md).

## 7. Known limits

The directory exposes names and usernames to every signed-in user: fine for a team server, but a public deployment would need a visibility setting (see D18 in [Design decisions](decisions.md)).

Email verification is implemented but **paused**, and `POST /auth/send-verify-email/:user_id` is still public (tracked as H-008 in the findings tracker). Docker Compose still carries database passwords in the file (H-051).

## 8. Reporting a vulnerability

Please do not open a public issue. Contact the maintainers privately and include steps to reproduce.
