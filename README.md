# AI Day Platform

AI Day has two React + TypeScript applications (`ui/` for customers and `admin/` for operations) and a Go API in `api/`.

## Phase 1 status

The customer UI and admin UI use Vite, React 19, TypeScript, ESLint, and Tailwind CSS 4. The Go API provides health, readiness, and public-configuration endpoints. Its PostgreSQL schema starts with user and identity foundations; later phases extend it with payments, credits, assets, and generation jobs.

## Prerequisites

* Node.js and npm
* Go 1.25.6+
* Docker Desktop (needed for PostgreSQL, Redis, and MinIO)

## Local setup

1. Copy `api/.env.example` to `api/.env` and set non-default provider values only locally.
2. Run `make setup`.
3. Start infrastructure with `make docker-up`.
4. In separate terminals, run `make api`, `cd ui && npm run dev`, and `cd admin && npm run dev`.

The API listens on `http://localhost:8080`. Available foundation endpoints are `GET /api/v1/health`, `GET /api/v1/ready`, and `GET /api/v1/config/public`.

`AUTH_MODE` supports `phone` and `microsoft`; only that safe selection and public product configuration are exposed to the frontend.

## Commands

`make test`, `make lint`, `make api`, `make worker`, `make docker-up`, and `make docker-down` are available. A migration runner and provider integrations are added in later phases.

## Security

Provider secrets, session secrets, and database credentials stay server-side. Never place them in a Vite `VITE_*` variable or commit `.env` files.