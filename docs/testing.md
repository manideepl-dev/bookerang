# Testing Bookerang

Use synthetic accounts and a disposable PostGIS database; never production or actual pilot data.

- Shared API: npm test from root.
- Backend: start disposable PostGIS with compose, export BOOKERANG_DB_PASSWORD, then run mvn test in backend. Tests exercise HTTP GraphQL, Flyway, authorization and concurrent acceptance.
- Web: npm install; npx playwright install chromium; npm run test -w web. Browser tests use a mocked GraphQL API.
- Native: npm run test -w mobile. Jest/React Native Testing Library exercises shared UI with mocked transport and location. This does not verify device-specific behavior.
- Build: npm run build -w web; from mobile run npx expo export --platform ios and npx expo export --platform android.

## Manual iOS and Android acceptance
With two synthetic accounts, verify location permission granted and denied, registration, sign-in, listing, discovery, request, acceptance, chat, blocking, reporting and return. Check cancellation, competing requests, app relaunch, permission revocation and poor network handling.

Tests have been committed as source. Execution and dependency compatibility remain unverified.
