# Repository Guidelines

## Project Structure & Module Organization

The monorepo splits server and client code: `backend/` hosts the Go services (clean architecture with `cmd/`, `internal/`, `domain/`, and Ent-generated `db/`), while `frontend/` contains the Vite + React app under `src/` with feature folders for `components/`, `pages/`, and `services/`. Shared documentation lives in `docs/`, infrastructure manifests in the root (`docker-compose*.yml`, `Makefile`), and generated assets stay out of source—keep build output in `tmp/` or feature-specific folders.

## Build, Test, and Development Commands

- `docker-compose -f docker-compose.dev.yml up -d`: spin up PostgreSQL and Redis for local work.
- `cd backend && go run cmd/main.go cmd/wire_gen.go`: launch the API at `http://localhost:8888` after syncing modules with `go mod tidy`.
- `cd frontend && npm install && npm run dev`: start the UI on `http://localhost:5175`; use `npm run build:prod` for production bundles.
- `make install-hooks`: install repo git hooks so pre-commit/pre-push run Go formatting, vet, lint, and front-end lint/build checks.

## Coding Style & Naming Conventions

Go code must be formatted with `gofmt`/`goimports`; prefer Go package names that mirror directory purpose (`user`, `general_agent`) and CamelCase exported identifiers. React/TypeScript follows ESLint + Prettier defaults (2-space indent, single quotes via config) and component filenames in PascalCase (`UserCard.tsx`), hooks in camelCase (`useSession.ts`). Keep environment files suffixed `.example` and avoid committing secrets.

## Testing Guidelines

Run `cd backend && go test ./...` for backend suites; add table-driven tests alongside implementation packages. Use `make test` if you prefer the aggregated target. Frontend tests are optional today; when adding them, wire them to `npm test` so CI hooks stay green. Name test files `*_test.go` and `<Component>.test.tsx` for consistency, and document unusual fixtures under `docs/`.

## Commit & Pull Request Guidelines

Adopt Conventional Commit prefixes observed in history (`feat:`, `refactor(auth):`, `docs:`). Scope optional but informative (`feat(ui-auth):`). Commits should be scoped and lint-clean before submission. PRs need a concise summary, linked issues or Jira tickets, test evidence (`go test`, `npm run lint`) inline, and screenshots for visible UI changes. Highlight data migrations or new env vars in the PR body so reviewers can sync their setups.

## Environment & Security Notes

Copy `.env.example` files rather than editing checked-in configs and keep credentials in local `.env`. When debugging AI integrations, prefer feature flags over hard-coded keys, and scrub logs before sharing. Remove any sample data from `tmp/` before pushing.

## Language and Localization Policy

The default working environment of this repository is Simplified Chinese (简体中文).
All documentation, logs, comments, commit messages, and user-facing interface text must use Simplified Chinese unless there is a specific reason to keep English (e.g., third-party API keywords or open-source license notices).

Key Requirements:
Internal documentation and comments → use Simplified Chinese.
Console output and logs → use Simplified Chinese for clarity in domestic deployment.
Frontend UI and backend API responses → default to Simplified Chinese, unless multi-language support is explicitly implemented.
Git commit messages → can use English prefixes (feat:, fix:), but the content should describe changes in Simplified Chinese.
