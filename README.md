# Bookerang

A small, invitation-only neighborhood book-lending pilot. This is a reconstructed development checkpoint, not the original lost archive.

## Architecture
Java 21 / Spring Boot / Spring GraphQL; PostgreSQL with PostGIS; shared JavaScript GraphQL transport. Web and native clients are planned next. See docs/v0.4-recovery.md for agreed scope.

## Local backend
Install Java 21, Maven and Docker. Set BOOKERANG_DB_PASSWORD to a development-only random secret and BOOKERANG_PILOT_INVITE to a separate secret of at least 16 characters. Run:

    docker compose up -d
    cd backend
    mvn spring-boot:run

The API listens on http://localhost:8080/graphql. Flyway installs the database schema. The default bind address is loopback. The shared API tests can be run with Node 20+ using:

    cd shared
    npm test

## Status
Schema, migration, GraphQL controllers and shared operations are committed. Build/runtime and database integration tests have not yet been run. Client UI is not implemented at this checkpoint. No production deployment, public registration or payment processing. Do not invite real users before integration/security testing, rate limiting, operation complexity limits, privacy review, account/session controls and a report-handling process.
