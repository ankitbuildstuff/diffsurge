# diffsurge

> Catch breaking API changes before your users do.

Diffsurge helps you detect breaking API changes before they reach production.

## Install

```bash
npm install -g diffsurge
```

This install path uses prebuilt binaries on supported platforms, so Go is not required.

Or use Docker:

```bash
docker run equixankit/diffsurge-cli --help
```

## Setup

1. **Create an API key** in the [Diffsurge dashboard](https://app.diffsurge.com) → Settings → API Keys
2. **Add it to your `.env` file** in your project root:

```env
SURGE_API_KEY=diffsurge_live_your_key_here
SURGE_PROJECT_ID=your-project-uuid
```

3. **Verify it works:**

```bash
surge whoami
```

## Commands

### `surge whoami`

Verify your API key is valid and see account info:

```bash
surge whoami
```

### `surge check`

Run API checks in your CI/CD pipeline:

```bash
# Basic check — validates API key, fetches traffic & schema stats
surge check --project-id abc-123

# With local schema diff — detect breaking changes before deploy
surge check --project-id abc-123 --schema openapi.yaml --fail-on-breaking
```

### `surge diff`

Compare two JSON API responses:

```bash
surge diff --old response-v1.json --new response-v2.json
```

### `surge schema diff`

Compare two OpenAPI/Swagger schema files and detect breaking changes:

```bash
surge schema diff --old api-v1.yaml --new api-v2.yaml --fail-on-breaking
```

### `surge replay`

Replay captured traffic against a target server:

```bash
surge replay --source captured.json --target http://localhost:8080
```

## CI/CD Integration

### GitHub Actions

```yaml
name: API Check
on: [push, pull_request]
jobs:
  api-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm install -g diffsurge
      - run: surge check --project-id ${{ secrets.SURGE_PROJECT_ID }}
        env:
          SURGE_API_KEY: ${{ secrets.SURGE_API_KEY }}
```

### GitLab CI

```yaml
api-check:
  image: node:20
  script:
    - npm install -g diffsurge
    - surge check --project-id $SURGE_PROJECT_ID
  variables:
    SURGE_API_KEY: $SURGE_API_KEY
```

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `SURGE_API_KEY` | Your API key (starts with `diffsurge_live_`) | — |
| `SURGE_API_URL` | API base URL | `https://api.diffsurge.com` |
| `SURGE_PROJECT_ID` | Default project ID | — |

All variables can also use the `DIFFSURGE_` prefix (e.g., `DIFFSURGE_API_KEY`) as a fallback.

## Flags

All commands support these global flags:

```
--api-key     API key (overrides SURGE_API_KEY)
--api-url     API base URL (overrides SURGE_API_URL)
--project-id  Project ID (overrides SURGE_PROJECT_ID)
```

## License

MIT
