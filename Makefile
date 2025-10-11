# WhaleHire Project Makefile
# Manages git hooks and project-wide tasks

.PHONY: install-hooks uninstall-hooks help

# Default target
help:
	@echo "WhaleHire Project Management"
	@echo "Available targets:"
	@echo "  install-hooks    - Install git hooks for code quality checks"
	@echo "  uninstall-hooks  - Remove git hooks"
	@echo "  help            - Show this help message"

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