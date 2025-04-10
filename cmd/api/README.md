# API Command

The `api` command allows you to make direct API requests to SailPoint Identity Security Cloud endpoints.

Similar to the GitHub CLI's `api` command, this provides a simple way to hit any custom API endpoint with custom headers, parameters, and content.

## Usage

### GET Request

```bash
sail api get /beta/accounts
sail api get /beta/accounts/123 --header "Accept: application/json" --pretty
sail api get /beta/identities --query "limit=100" --query "offset=0" --output identities.json
sail api get /beta/accounts --jsonpath "$.items[*].id"  # Extract all account IDs
```

### POST Request

```bash
sail api post /beta/accounts --body '{"accountId":"test123", "name":"Test Account"}'
sail api post /beta/import/data --body-file request.json --header "Content-Type: application/json"
```

### PUT Request

```bash
sail api put /beta/accounts/123 --body '{"name":"Updated Account"}'
```

### PATCH Request

```bash
sail api patch /beta/accounts/123 --body '{"name":"Patched Account"}'
```

### DELETE Request

```bash
sail api delete /beta/accounts/123
```

## Common Options

All commands support the following options:

- `--header`, `-H`: Set HTTP headers (can be used multiple times, format: 'Key: Value')
- `--output`, `-o`: Output file to save the response (if not specified, prints to stdout)
- `--pretty`, `-p`: Pretty print JSON response
- `--jsonpath`, `-j`: JSONPath expression to evaluate on the response

### Body Options (POST, PUT, PATCH)

- `--body`, `-b`: Request body content as a string
- `--body-file`, `-f`: File containing the request body
- `--content-type`, `-c`: Content type of the request body (default: application/json)

### Query Options (GET, DELETE)

- `--query`, `-q`: Query parameters (can be used multiple times, format: 'key=value')

## JSONPath Examples

The `--jsonpath` option allows you to extract specific values from JSON responses using JSONPath expressions:

```bash
# Extract a single value
sail api get /beta/accounts/123 --jsonpath "$.name"

# Extract nested values
sail api get /beta/accounts/123 --jsonpath "$.attributes.email"

# Extract array elements
sail api get /beta/accounts --jsonpath "$.items[*].id"

# Extract multiple values
sail api get /beta/accounts/123 --jsonpath "$.name" --jsonpath "$.id"
```

JSONPath expressions follow the standard JSONPath syntax:
- `$`: Root object
- `.`: Child operator
- `[]`: Array access
- `[*]`: Array wildcard
- `..`: Recursive descent 