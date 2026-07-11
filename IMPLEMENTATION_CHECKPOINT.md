# AI Day implementation checkpoint

Last updated: 2026-07-11

## Working rule

Implementation is currently proceeding without running tests, linting, builds, or smoke checks at the user’s request.

## Completed foundations

* Vite React customer and admin shells with Tailwind, routing, and customer public-config query.
* Go API health, readiness, public config, structured logging, and graceful shutdown.
* PostgreSQL migrations for users, sessions, OTP records, payments, credits, assets, jobs, webhooks, and audit logs.
* Session token hashing, `auth/me`, logout, and authentication-mode-gated routes.
* Credit balance endpoint and an atomic credit-reservation plus generation-job transaction.
* Interfaces for AI and object-storage providers, Razorpay HMAC verification, and comic-pipeline models.

## In progress

* Real MSG91 and Microsoft authentication flows.
* Razorpay order/webhook persistence.
* S3/MinIO signed uploads and ownership completion checks.
* Worker execution, OpenAI image generation, comic composition/PDF output.
* Admin authorization and audited admin actions.

## Required external configuration

* PostgreSQL, Redis, and MinIO/Docker Desktop.
* MSG91 credentials and template ID.
* Microsoft Entra application credentials and redirect URI.
* Razorpay key ID, secret, and webhook secret.
* S3-compatible storage credentials.
* OpenAI API key and selected image/model configuration.

## Verification backlog

* Run migrations on a fresh PostgreSQL database.
* Run Go tests/vet, frontend lint/build, and end-to-end payment/upload/generation checks.
* Review generated user-facing and admin UI in a browser.

## Latest implementation

* Added server-owned upload intents for permitted selfie MIME types and size limits.
* Added source-asset ownership enforcement before either generation can be queued.
* Added backend admin authorization middleware using the session's administrator flag.
* Storage signing remains blocked on configured S3/MinIO provider credentials and client implementation.

## Phase status (2026-07-11)

| Phase | Status | Exit-condition blocker |
| --- | --- | --- |
| 1. Foundation | Partial | Docker/PostgreSQL migration run still pending |
| 2. Authentication | Partial | MSG91 and Microsoft provider flows require configured credentials |
| 3. Razorpay & credits | Partial | Razorpay order/webhook client and secrets still required |
| 4. Asset storage | Partial | S3/MinIO signed-URL provider still required |
| 5. PixArt | Partial | Queue, OpenAI provider, output storage still required |
| 6. Comic | Partial | Orchestrator, panel renderer, PDF compositor still required |
| 7. Admin | Partial | Admin endpoints/dashboard data and mutations still required |
| 8. Hardening | Not started | Requires completed flows and verification pass |

Added queue, CSRF, and audit-log primitives in the current implementation pass.
