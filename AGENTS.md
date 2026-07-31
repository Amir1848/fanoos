# Agent Instructions — Marketino

This file is the entry point for AI coding agents working in this repository.

## Start Here

Before making changes, read the system architecture doc:

- [`docs/architucture.md`](docs/architucture.md) — components, data flow, deployment (Docker Compose), and database (PostgreSQL) design.

That document describes what each part of the system does and where it lives in the repo. Keep it up to date: if you add, remove, or change the responsibility of a component (backend, frontend, broker, firmware, database), update `docs/architucture.md` in the same change.

## Repo Layout

- `/marketino` — Go backend server.
- `/new-angular-project` — Angular frontend.
- `/docs` — architecture and feature documentation.

## Ground Rules

- Backend persistence goes through PostgreSQL; schema changes belong in `/marketino/scripts/migrations`.
- ESP32 devices communicate with the backend only via the MQTT broker, never directly.
