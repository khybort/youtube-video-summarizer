# Testing Guide

## Overview

This project includes comprehensive testing for both backend and frontend components. All tests must pass before code can be committed.

## Running Tests

### Run All Tests
```bash
make test
```

This runs both backend and frontend tests.

### Run Backend Tests Only
```bash
make test-backend
```

Runs all backend unit and integration tests:
- Package tests (`pkg/...`)
- Similarity service tests
- Middleware tests
- Handler tests

### Run Frontend Tests Only
```bash
make test-frontend
```

Runs Playwright E2E tests for the frontend.

### Run All Backend Tests (Including Integration)
```bash
make test-all
```

## Pre-commit Hook

A pre-commit hook is installed that automatically runs all tests before allowing a commit. If any test fails, the commit will be blocked.

### How It Works

1. When you run `git commit`, the pre-commit hook automatically executes
2. It runs `make test-backend` to check backend tests
3. It runs `make test-frontend` to check frontend tests
4. If all tests pass, the commit proceeds
5. If any test fails, the commit is blocked and you'll see an error message

### Bypassing Pre-commit Hook (Not Recommended)

If you absolutely need to bypass the pre-commit hook (e.g., for WIP commits), you can use:

```bash
git commit --no-verify
```

**Warning**: Only use this in exceptional circumstances. The pre-commit hook ensures code quality.

### Manual Pre-commit Hook Execution

You can manually test the pre-commit hook:

```bash
.git/hooks/pre-commit
```

## Test Structure

### Backend Tests

- **Location**: `backend/`
- **Framework**: Go's built-in testing + `testify`
- **Coverage**:
  - Package tests (pricing, LLM providers)
  - Service tests (similarity, cost, summary, transcript, embedding)
  - Handler tests (video, cost)
  - Middleware tests (CORS, request ID)

### Frontend Tests

- **Location**: `frontend/e2e/`
- **Framework**: Playwright
- **Coverage**:
  - Navigation tests
  - Video actions (add, delete, view)
  - Video detail page
  - Transcript viewer
  - Summary display
  - Similar videos
  - Cost analysis
  - Settings
  - Search functionality
  - Dashboard
  - Integration tests

## CI/CD Integration

The pre-commit hook ensures that:
- All tests pass locally before committing
- Code quality is maintained
- Broken code is not committed to the repository

For CI/CD pipelines, the same test commands can be used:
- `make test-backend` for backend tests
- `make test-frontend` for frontend tests

## Troubleshooting

### Pre-commit Hook Not Running

If the pre-commit hook is not executing:

1. Check if the file exists and is executable:
   ```bash
   ls -la .git/hooks/pre-commit
   ```

2. Make it executable:
   ```bash
   chmod +x .git/hooks/pre-commit
   ```

### Tests Failing Locally

If tests are failing:

1. Check that all dependencies are installed
2. Ensure services are running (for integration tests)
3. Check test logs for specific error messages
4. Run tests individually to isolate issues:
   ```bash
   make test-backend
   make test-frontend
   ```

### Frontend Tests Requiring Services

Frontend E2E tests require:
- Frontend dev server running (automatically started by Playwright)
- Backend API available (if testing API integration)

The Playwright config automatically starts the frontend dev server, but you may need to ensure the backend is running for full integration tests.

