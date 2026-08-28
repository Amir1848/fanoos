# Build context is the REPO ROOT (same as the main Dockerfile), because the
# SQL migrations live under fanoos/scripts/migrations:
#   docker build -f migrator.Dockerfile -t amir1848/fanoos-migrator .
# Published by .github/workflows/docker-publish.yml alongside the backend.

FROM docker.arvancloud.ir/migrate/migrate:v4.18.2

# Baked in so the VPS needs no repo checkout: the image is self-contained.
COPY fanoos/scripts/migrations /migrations

# Entrypoint of the base image is `migrate`; callers only pass args:
#   docker run amir1848/fanoos-migrator -path=/migrations/ -database=postgres://... up
