# Agent Instructions — Fanoos

This file is the entry point for AI coding agents working in this repository.

## Start Here

Before making changes, read:

- [`docs/README.md`](docs/README.md) — project vision, goals, and design principles (why this exists, what it must never break). Includes an "AI Context" section with constraints to assume in every change.
- [`docs/architucture.md`](docs/architucture.md) — components, data flow, deployment (Docker Compose), and database (PostgreSQL) design.

The architecture doc describes what each part of the system does and where it lives in the repo. Keep it up to date: if you add, remove, or change the responsibility of a component (backend, frontend, broker, firmware, database), update `docs/architucture.md` in the same change.

## Repo Layout

- `/fanoos` — Go backend server.
- `/new-angular-project` — Angular frontend.
- `/docs` — architecture and feature documentation.
- `/compose.yml` + `/mosquitto` — Docker Compose stack and Mosquitto (MQTT broker) config.

## Ground Rules

- Backend persistence goes through PostgreSQL; schema changes belong in `/fanoos/scripts/migrations`.
- ESP32 devices communicate with the backend only via the MQTT broker, never directly.
