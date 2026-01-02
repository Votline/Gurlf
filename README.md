# Gurlf

Gurlf is a custom configuration format and Go parsing library designed as an alternative to JSON in cases where escaping becomes a problem.

It was created to handle complex, real-world configurations (HTTP headers, cookies, request bodies, embedded JSON) without breaking readability or data integrity.

---

## Motivation

Gurlf was born from practical issues encountered while developing [gurl-cli](https://github.com/Votline/gurl-cli).

When adding support for cookies and complex headers, nested JSON structures had to be embedded inside JSON configs. This led to excessive escaping, unreadable configs and broken data after parsing.

Gurlf solves this by:
- Removing escaping entirely
- Treating values as raw data
- Supporting multiline values natively

---

## Format Overview

Gurlf uses a simple, section-based format:
```bash
[config_name]
KEY: value
MULTILINE_KEY: `
    multiline
    value
    without escaping
`
[\config_name]
```

### Syntax Rules

- `[name]` — start of a config section
- `[\name]` — end of a config section
- `KEY: value` — key-value pair
- Backticks (`) allow multiline values
- No escaping rules — data is read as-is

---

## Example

```bash
[request_1]
ID: 1
HEADERS: Content-Type: application/json
BODY: `
    { "key": "value", "nested": { "data": "here" } }
`
COOKIE: session=abc123; path=/; domain=example.com
[\request_1]
```
---

## Features

- Zero escaping — raw data is preserved
- Multiline values via backticks
- Human-readable and writable format
- Multiple config sections per file
- Type-safe unmarshalling into Go structs
- Designed for HTTP-related configurations

---

## Usage

### As a Go Library

```go
import "github.com/Votline/Gurlf"

// Scan config file
data, err := gurlf.ScanFile("config.gurlf")
if err != nil {
    log.Fatal(err)
}

// Unmarshal section into struct
type Config struct {
    ID      int    `gurlf:"ID"`
    Body    string `gurlf:"BODY"`
    Headers string `gurlf:"HEADERS"`
    Cookie  string `gurlf:"COOKIE"`
    Name    string `gurlf:"config_name"` //tag for the config name without brackets: [config_name]
}

var cfg Config
err = gurlf.Unmarshal(data[0], &cfg)
if err != nil {
    log.Fatal(err)
}
```

---

## When to Use Gurlf

- HTTP request definitions (headers, body, cookies)
- Configs containing embedded JSON
- Any scenario where JSON escaping hurts readability
- Tooling configs where humans write and edit files
- Prototyping request-based workflows

---

## Tech Stack

- Go 1.25+
- reflect (for unmarshalling)
- uber-go/zap (logging)

---

## Design Goals

- Simplicity over features
- Predictable parsing
- No hidden transformations
- Minimal syntax
- Developer-friendly ergonomics

---

## License

- **License:** This project is licensed under  [MIT](LICENSE)
- **Third-party Licenses:** The full license texts are available in the  [licenses/](licenses/)
