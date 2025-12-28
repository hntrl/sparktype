# Remote Specs

This guide covers fetching OpenAPI specifications from remote URLs, including authentication, environment variables, and best practices.

## Basic Remote Spec

Fetch a spec from a URL:

```jsonc
{
  "specs": {
    "api": {
      "url": "https://api.example.com/openapi.json"
    }
  }
}
```

sparktype detects the format from:
1. `Content-Type` response header
2. URL file extension (`.yaml`, `.yml`, `.json`)

## Authentication

Many APIs require authentication to access their OpenAPI specs.

### Bearer Token

```jsonc
{
  "specs": {
    "api": {
      "url": "https://api.example.com/openapi.json",
      "headers": {
        "Authorization": "Bearer ${API_TOKEN}"
      }
    }
  }
}
```

### API Key

```jsonc
{
  "specs": {
    "api": {
      "url": "https://api.example.com/openapi.json",
      "headers": {
        "X-API-Key": "${API_KEY}"
      }
    }
  }
}
```

### Basic Auth

```jsonc
{
  "specs": {
    "api": {
      "url": "https://api.example.com/openapi.json",
      "headers": {
        "Authorization": "Basic ${BASIC_AUTH_TOKEN}"
      }
    }
  }
}
```

Generate the token: `echo -n "username:password" | base64`

### Multiple Headers

```jsonc
{
  "specs": {
    "api": {
      "url": "https://api.example.com/openapi.json",
      "headers": {
        "Authorization": "Bearer ${API_TOKEN}",
        "X-Client-ID": "${CLIENT_ID}",
        "Accept": "application/json"
      }
    }
  }
}
```

## Environment Variables

Header values support `${VAR_NAME}` interpolation:

```jsonc
"headers": {
  "Authorization": "Bearer ${GITHUB_TOKEN}"
}
```

### Setting Variables

**Shell:**

```sh
export API_TOKEN="your-token-here"
sparktype generate
```

**Inline:**

```sh
API_TOKEN="your-token" sparktype generate
```

**dotenv (with shell):**

```sh
source .env && sparktype generate
```

### Missing Variables

If a variable is not set, the literal `${VAR_NAME}` string is kept. This usually causes authentication failures with a clear error.

## Common API Providers

### GitHub

```jsonc
{
  "specs": {
    "github": {
      "url": "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json",
      "headers": {
        "Authorization": "token ${GITHUB_TOKEN}"
      }
    }
  }
}
```

### Stripe

```jsonc
{
  "specs": {
    "stripe": {
      "url": "https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json"
    }
  }
}
```

Stripe's spec is public, no auth needed.

### Private GitHub Repo

```jsonc
{
  "specs": {
    "internal": {
      "url": "https://raw.githubusercontent.com/myorg/api-specs/main/openapi.yaml",
      "headers": {
        "Authorization": "token ${GITHUB_TOKEN}"
      }
    }
  }
}
```

### GitLab

```jsonc
{
  "specs": {
    "api": {
      "url": "https://gitlab.com/api/v4/projects/12345/repository/files/openapi.yaml/raw?ref=main",
      "headers": {
        "PRIVATE-TOKEN": "${GITLAB_TOKEN}"
      }
    }
  }
}
```

### AWS API Gateway

```jsonc
{
  "specs": {
    "api": {
      "url": "https://apigateway.us-east-1.amazonaws.com/restapis/abc123/stages/prod/exports/oas30",
      "headers": {
        "Authorization": "${AWS_AUTH_HEADER}"
      }
    }
  }
}
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Generate types
  env:
    API_TOKEN: ${{ secrets.API_TOKEN }}
  run: npx sparktype generate
```

### GitLab CI

```yaml
generate-types:
  variables:
    API_TOKEN: $API_TOKEN  # From CI/CD settings
  script:
    - npx sparktype generate
```

### Environment File

For local development, use a `.env` file:

```sh
# .env
API_TOKEN=your-development-token
```

Add to `.gitignore`:

```
.env
```

## Caching Remote Specs

Remote specs are fetched on every run. For faster builds:

### Local Cache Script

```sh
#!/bin/bash
# scripts/fetch-specs.sh

mkdir -p .specs

curl -H "Authorization: Bearer $API_TOKEN" \
  https://api.example.com/openapi.json \
  -o .specs/api.json
```

Update config to use local cache:

```jsonc
{
  "specs": {
    "api": {
      "path": "./.specs/api.json"
    }
  }
}
```

### CI Cache

```yaml
# GitHub Actions
- uses: actions/cache@v4
  with:
    path: .specs
    key: specs-${{ hashFiles('scripts/fetch-specs.sh') }}
    
- name: Fetch specs
  if: steps.cache.outputs.cache-hit != 'true'
  run: ./scripts/fetch-specs.sh
```

## Error Handling

### Connection Errors

```
failed to load spec: Get "https://api.example.com/openapi.json": dial tcp: lookup api.example.com: no such host
```

**Solutions:**
- Check URL is correct
- Verify network connectivity
- Check DNS resolution

### Authentication Errors

```
failed to load spec: unexpected status 401
```

**Solutions:**
- Verify token is correct
- Check token hasn't expired
- Ensure environment variable is set

### Invalid Spec

```
failed to load spec: yaml: unmarshal error
```

**Solutions:**
- Verify URL returns OpenAPI spec (not HTML error page)
- Check the Content-Type header
- Try downloading manually to inspect

## Security Best Practices

### Never Commit Secrets

```jsonc
// WRONG - token in config
"headers": {
  "Authorization": "Bearer ghp_xxxxxxxxxxxx"
}

// CORRECT - environment variable
"headers": {
  "Authorization": "Bearer ${GITHUB_TOKEN}"
}
```

### Use Read-Only Tokens

Create tokens with minimal permissions:
- GitHub: `repo` (read) or fine-grained with `contents: read`
- GitLab: `read_repository` scope

### Rotate Tokens Regularly

Set up token rotation schedules for production environments.

### Audit Access

Log which tokens access your specs:
- API Gateway request logging
- Token usage monitoring

## Fallback to Local

For reliability, consider maintaining local copies:

```jsonc
{
  "specs": {
    // Primary: remote spec
    "api": {
      "url": "https://api.example.com/openapi.json",
      "headers": { "Authorization": "Bearer ${API_TOKEN}" }
    }
    
    // Alternative config for offline development:
    // "api": {
    //   "path": "./specs/api.json"
    // }
  }
}
```

## Troubleshooting

### Test URL Manually

```sh
curl -H "Authorization: Bearer $API_TOKEN" \
  https://api.example.com/openapi.json \
  -o /dev/null -w "%{http_code}"
```

### Debug Environment Variables

```sh
# Check if variable is set
echo $API_TOKEN

# Run with debug output
API_TOKEN="test" sparktype validate
```

### Check Response Content

```sh
curl -H "Authorization: Bearer $API_TOKEN" \
  https://api.example.com/openapi.json | head -20
```

Verify it's actual OpenAPI content, not an error page.

## See Also

- [Configuration: Specs](/configuration/specs)
- [CI/CD Integration](./ci-cd)
- [Multiple Specs](./multi-spec)

