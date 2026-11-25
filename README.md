# Bookerang

A small, invitation-only neighborhood book-lending pilot. Reconstructed development source, not the original lost archive.

## Architecture
Java 21 / Spring Boot / Spring GraphQL; PostgreSQL/PostGIS; Next.js web using React Native Web; Expo/React Native iOS and Android. Shared UI and GraphQL transport live in shared/. See docs/v0.4-recovery.md for agreed scope.

## Local backend
Install Java 21, Maven, Docker and Node 20+. Export BOOKERANG_DB_PASSWORD and BOOKERANG_PILOT_INVITE (separate random development secrets; invite at least 16 characters). Run from root:

    docker compose up -d
    cd backend
    mvn spring-boot:run

The API listens on http://localhost:8080/graphql. Flyway installs the database schema. The default bind address is loopback.

## Clients
From the repository root:

    npm install
    npm run web

Open http://localhost:3000. For native testing set EXPO_PUBLIC_API_URL to the backend GraphQL URL and run npm run mobile. For physical devices on a trusted development network only, configure BOOKERANG_BIND_ADDRESS=0.0.0.0 on the backend and use the development machine LAN address in EXPO_PUBLIC_API_URL. Android emulator may use 10.0.2.2. Do not expose this HTTP development service to the internet. Production requires HTTPS and appropriate CORS configuration. Location is requested on demand, not continuously.

## Verify
Run npm test for mocked shared API tests, mvn test in backend, npm run build -w web and test Expo on both device platforms. These commands have not been executed in the reconstruction environment. Dependency versions and bundler compatibility require verification.

Use two invited test accounts: register, list a book, discover, request, accept, exchange messages, mark returned, rediscover. Test decline, cancel, archive, block and report; check unauthorized access and simultaneous accept/block/send operations. Reports are stored for manual review in the database.

## Limitations
Source implementations for all three clients and backend are committed but compilation, dependency installation, database integration and device tests remain pending. UI is shared, not separately polished for each platform. Sessions are held in client memory; sign-out clears local state but does not revoke the backend token. Chat and requests refresh manually; no subscriptions, push notifications, payment processing or continuous tracking. Before real users: rate limiting, query complexity limits, security and privacy review, account/session controls, moderation process and HTTPS. No production deployment or app-store readiness claimed.
