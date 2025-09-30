# Repository Guidelines

## Project Structure & Module Organization
WhaleHire is split across `backend/` (Go) and `ui/` (Next.js). The backend follows clean architecture: entrypoints live in `backend/cmd`, application flow under `backend/internal/{module}/{handler|usecase|repo}`, and shared packages in `backend/pkg`. Ent-generated schemas stay in `backend/ent` and migrations under `backend/migration`. The front end keeps all route logic in `ui/src/app`, reusable components under `ui/src/components`, and shared helpers in `ui/src/lib`. API references and onboarding notes live in `docs/`. Root-level Docker Compose files provision Postgres and Redis for local development.

## Build, Test, and Development Commands
- `make install-hooks` installs repo-wide Git hooks for Go and UI linting.
- `docker-compose up -d whalehire-db whalehire-redis` boots required services.
- `cd backend && go run cmd/main.go cmd/wire_gen.go` starts the API at `:8888`.
- `cd backend && go test ./...` runs the Go unit and integration suite.
- `cd ui && npm run dev` launches the Next.js dev server on `:3000`.
- `cd ui && npm run lint` enforces ESLint and TypeScript checks before shipping.

## Coding Style & Naming Conventions
Go code must remain gofmt/goimports clean; prefer idiomatic naming (`AdminService`, `NewUserRepo`). Keep handler/usecase/repo files within matching module folders and suffix tests with `_test.go`. The UI uses 2-space indentation, TypeScript strict mode, and Tailwind utility classes. Co-locate UI components in `ui/src/components/<area>/ComponentName.tsx` using PascalCase, while helpers in `ui/src/lib` stay kebab-case. Run `golangci-lint` and `npm run lint` before requesting review.

## Testing Guidelines
Favor table-driven Go tests alongside their packages (`backend/internal/user/...`). Use `enttest` fixtures where data access is involved and clean up with transactions. UI behavior checks should live under `ui/src/__tests__/` using Playwright or Jest-compatible tooling; stub API calls via MSW or fetch mocks. Target meaningful coverage on new features and document manual verification steps (API endpoints hit, UI flows exercised) in the pull request if automated tests are impractical.

## Commit & Pull Request Guidelines
Follow the existing Conventional Commit style, e.g., `feat(session): support admin logout (#5)` or `fix(docker): adjust backend URL`. Keep messages in English with concise scopes. Each PR should include: a one-paragraph summary, linked issue or ticket, screenshots/recordings for UI changes, and a checklist of tests run (`go test`, `npm run lint`, manual QA). Request reviews by module owners (`backend`, `ui`) and wait for CI to finish before merging.

## Configuration & Security Tips
Copy `.env.example` to `.env` for local overrides and never commit secrets. Backend configuration lives in `backend/config`, while frontend environment variables belong in `ui/.env.local` with the `NEXT_PUBLIC_` prefix when exposed to the browser. Use the provided Docker Compose files for local stacks and refresh credentials before demos or sharing datasets.
