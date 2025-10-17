# WhaleHire Project Makefile
# Manages git hooks and project-wide tasks

.PHONY: install-hooks uninstall-hooks quality-check quality-check-frontend quality-check-backend help

# Default target
help:
	@echo "WhaleHire Project Management"
	@echo "Available targets:"
	@echo "  install-hooks           - Install git hooks for code quality checks"
	@echo "  uninstall-hooks         - Remove git hooks"
	@echo "  quality-check           - Run code quality checks for both backend and frontend"
	@echo "  quality-check-frontend  - Run only frontend quality checks"
	@echo "  quality-check-backend   - Run only backend quality checks"
	@echo "  help                   - Show this help message"

# Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@mkdir -p .git/hooks
	@chmod +x .githooks/pre-commit
	@cp .githooks/pre-commit .git/hooks/pre-commit
	@echo "Git hooks installed successfully!"
	@echo "  - pre-commit: runs code quality checks before commit"
	@echo ""
	@echo "Hooks will check:"
	@echo "  - Backend Go files: formatting, vet, lint, mod tidy"
	@echo "  - Frontend files: ESLint, TypeScript check, build"

# Uninstall git hooks
uninstall-hooks:
	@echo "Uninstalling git hooks..."
	@rm -f .git/hooks/pre-commit
	@echo "Git hooks uninstalled successfully!"

# Run code quality checks
quality-check:
	@echo "=========================================="
	@echo "Running Code Quality Checks"
	@echo "=========================================="
	@echo ""
	@echo "1. Backend Quality Checks..."
	@echo "------------------------------------------"
	@if command -v goimports >/dev/null 2>&1; then \
		cd backend && $(MAKE) check || exit 1; \
	else \
		echo "⚠️  Warning: goimports not found. Skipping backend checks."; \
		echo "   Install with: cd backend && make install-tools"; \
	fi
	@echo ""
	@echo "2. Frontend Quality Checks..."
	@echo "------------------------------------------"
	@cd frontend && $(MAKE) quality-check || exit 1
	@echo ""
	@echo "=========================================="
	@echo "✅ Quality checks completed!"
	@echo "=========================================="

# Run only frontend quality checks
quality-check-frontend:
	@echo "Running Frontend Quality Checks..."
	@cd frontend && $(MAKE) quality-check

# Run only backend quality checks
quality-check-backend:
	@echo "Running Backend Quality Checks..."
	@cd backend && $(MAKE) check