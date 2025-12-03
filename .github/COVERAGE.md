# Coverage Badges Setup

This project includes automated test coverage tracking using GitHub Actions and Codecov.

## What's Included

### Badges in README.md

1. **Tests Badge**: Shows if tests are passing on the main branch
2. **Codecov Badge**: Shows current test coverage percentage
3. **Go Report Card**: Code quality score
4. **Go Version**: Minimum Go version required
5. **License**: Project license (MIT)

### GitHub Actions Workflow

File: `.github/workflows/test.yml`

The workflow:
- Runs on every push to `main` and on all pull requests
- Sets up PostgreSQL and Valkey services for integration tests
- Runs all tests with race detection enabled
- Generates coverage report
- Uploads coverage to Codecov
- Runs golangci-lint for code quality

## Setup Instructions

### 1. Enable GitHub Actions

GitHub Actions is enabled by default for public repositories. For private repos, enable it in Settings → Actions.

### 2. Setup Codecov (Optional but Recommended)

1. Go to [codecov.io](https://codecov.io)
2. Sign in with your GitHub account
3. Add your repository
4. Get your upload token from Settings
5. Add the token as a GitHub secret:
   - Go to your repository Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `CODECOV_TOKEN`
   - Value: Your Codecov upload token

**Note**: For public repositories, the token is optional. Coverage will still be tracked without it.

### 3. Setup Go Report Card (Automatic)

Go Report Card automatically scans public Go repositories on GitHub. Just visit:
```
https://goreportcard.com/report/github.com/pguia/iam
```

The first visit will trigger an initial scan. The badge will update automatically.

## Local Coverage Testing

Generate coverage report locally:

```bash
# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

## Current Coverage

As of the latest test run:
- **Service Package**: 47.4%
- **Overall**: Will be calculated by Codecov once workflow runs

## Coverage Goals

- Maintain minimum 70% coverage for critical packages (service, domain)
- Aim for 80%+ coverage on permission evaluation logic
- 90%+ coverage on cache implementations

## Troubleshooting

### Badges Not Showing

1. **Tests Badge**: Verify GitHub Actions workflow has run at least once
2. **Codecov Badge**: Ensure at least one successful coverage upload
3. **Go Report Card**: Visit the URL manually to trigger first scan

### Coverage Not Uploading

- Check that `CODECOV_TOKEN` secret is set correctly
- Verify the workflow has permission to upload artifacts
- Check GitHub Actions logs for upload errors

### Workflow Failing

- Ensure PostgreSQL and Valkey services start successfully
- Check that all required environment variables are set
- Verify Go version compatibility
