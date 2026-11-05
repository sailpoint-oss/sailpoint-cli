# API Command

The `api` command allows you to make direct API requests to SailPoint Identity Security endpoints.

Similar to the GitHub CLI's `api` command, this provides a simple way to hit any custom API endpoint with custom headers, parameters, and content.

## Usage

### Stateless automation request

`sail api request METHOD /endpoint` uses exactly the complete `SAIL_BASE_URL`,
`SAIL_CLIENT_ID` and `SAIL_CLIENT_SECRET` environment triple. It does not read or
write CLI configuration, use the keyring/token cache, prompt, retry, or follow
redirects. Supply credentials through the execution account's existing secret
store/environment, never as command arguments. It requests a fresh PAT access
token per operation. This extra token request is a deliberate alternative to
reusing an environment cache whose execution identity may differ.

```bash
sail api request GET '/v3/sources?limit=1' --timeout 120s
sail api request POST '/beta/sources/source-id/load-accounts' \
  --body-file multipart.body --content-type 'multipart/form-data; boundary=my-boundary' \
  --timeout 600s
```

The body file is streamed with its exact bytes and content length. Multipart
construction remains the caller's responsibility. `--header 'Key: Value'`
accepts additional headers, except authentication, cookie, host and framing
overrides. `--env` and `--debug` are refused. The timeout covers authentication
and the request. HTTPS is required except for HTTP loopback test servers.

Stdout contains exactly one `sail-api-response/1` JSON envelope:

```json
{"format":"sail-api-response/1","status":202,"headers":{"Content-Type":["application/json"]},"body":"eyJpZCI6InRhc2staWQifQ=="}
```

`body` is base64 for the raw response bytes, including non-JSON error pages.
An HTTP refusal still exits zero; callers must evaluate `status`. Process
failure, an incomplete envelope or a deadline after a write does not establish
whether the write occurred. Never retry an uncertain mutation. Headers include
`Retry-After` and `X-Total-Count` where supplied. Authentication/transport errors
have fixed diagnostics without credential values or response/request dumps.
An authentication HTTP refusal also returns its status/headers, with a fixed
body identifying the authentication phase and `requestSubmitted: false`;
the token endpoint's raw response is never reflected into the envelope.
The existing get/post/put/patch/delete commands retain their behavior.

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

The legacy get/post/put/patch/delete commands support the following options;
the stateless request command's separate options are described above:

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
