# Gurlf

**Gurlf** is a high-performance, escaping-free configuration format and Go parsing library.

It is engineered to solve the "embedded data" problem: handling complex, multiline payloads (like HTTP bodies, SQL queries, or nested JSON) without the readability nightmare of escape sequences.

> **Why check this out?** This library implements a custom **zero-allocation lexer** and uses low-level reflection to map raw bytes directly to Go structs with minimal overhead.

---

## ⚡ Motivation: The Escaping Hell

Gurlf was born from the practical needs of [gurl-cli](https://github.com/Votline/gurl-cli). When defining HTTP requests that contain JSON bodies or complex Cookies, standard JSON configurations become unreadable.

**The Problem (JSON):**

```json
{
  "id": 1,
  "body": "{\n  \"query\": \"Select * FROM users WHERE name = \\\"John\\\"\"\n}"
}

```

**The Solution (Gurlf):**

```bash
[request]
ID: 1
BODY: `
    {
      "query": "Select * FROM users WHERE name = \"John\""
    }
`
[\request]

```

*No escaping. Raw data remains raw.*

---

## 🚀 Key Features

* **High Performance**
* **Extremely low latency** parsing logic.
* **Zero-Escaping:** Values are treated as raw byte streams. What you see is what you get.
* **Smart Multiline Support:** Handles nested raw strings via intelligent delimiter detection (see *Syntax*).
* **Custom handwritten scanner** (no regex).
* **Zero allocations** during the Unmarshal phase (reusing internal buffers).
* **Struct Mapping:** Type-safe unmarshalling into Go structs using `reflect`.

---

## 📝 Syntax & Smart Nesting

Gurlf uses a section-based format.

* Sections start with `[name]` and end with `[\name]`.
* Keys are defined as `KEY: value`.

### The "Smart Backtick" System

To support configurations within configurations (or embedding code blocks that contain backticks), Gurlf uses a strict newline-based delimiter rule.

A multiline block is **only** closed if the backtick is isolated on its own line (surrounded by newlines: \n`\n).

This allows you to embed backticks inside your values freely, as long as they are inline.

**Example: Config inside a Config**

```bash
[outer_config]
Type: wrapper
# The parser ignores the inline backticks inside the block
# because they are not isolated on a new line.
Payload: `
    [inner_config]
    Title: "Embedded Config"
    JSON_Body: `{ "key": "value" }`   <-- These backticks do not close the block
    [\inner_config]
`
[\outer_config]

```

---

## 📊 Benchmarks

Gurlf is optimized for speed and memory efficiency. The core unmarshaller achieves **zero allocations** per operation by leveraging `unsafe` string casting and efficient buffer pooling.

**Environment:** AMD Ryzen 7 5800U, Linux/amd64.

| Benchmark | Iterations | Time (ns/op) | Bytes/op | Allocs/op |
| --- | --- | --- | --- | --- |
| **Core Unmarshal** | **22,315,796** | **48.19 ns/op** | **0 B/op** | **0 allocs/op** |
| Core Marshal | 5,601,400 | 203.20 ns/op | 144 B/op | 2 allocs/op |
| **Scanner Scan** | 2,931,316 | 387.50 ns/op | 160 B/op | 4 allocs/op |
| Scanner Emit | 26,475,750 | 44.56 ns/op | 0 B/op | 0 allocs/op |
| Scanner FindStart | 139,339,255 | 8.64 ns/op | 0 B/op | 0 allocs/op |
| Scanner FindEnd | 139,339,255 | 8.63 ns/op | 0 B/op | 0 allocs/op |
| Scanner FindKeyValue | 78,267,078 | 13.25 ns/op | 0 B/op | 0 allocs/op |


*Note: The scanner allocates minimal memory only for the slice headers of the returned entries, keeping garbage collection pressure negligible.*

---

## 💻 Usage

### Installation

```bash
go get github.com/Votline/Gurlf

```

### Go Example

```go
package main

import (
	"fmt"
	"log"
	"github.com/Votline/Gurlf"
)

// Define your struct with `gurlf` tags
type Config struct {
	ID      int    `gurlf:"ID"`
	Headers string `gurlf:"HEADERS"`
	Body    string `gurlf:"BODY"`
	Section string `gurlf:"config_name"` // Captures the [section_name]
}

func main() {
	// Scan the file
	sections, err := gurlf.ScanFile("config.gurlf")
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over sections
	for _, sectionData := range sections {
		var cfg Config
		
		// Unmarshal data into struct (Zero allocation)
		if err := gurlf.Unmarshal(sectionData, &cfg); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Loaded [%s]: ID=%d\n", cfg.Section, cfg.ID)
	}
}

```

---

## 🛠 Tech Stack

* **Language:** Go 1.25+
* **Core:** `reflect`, `unsafe`, `sync.Pool`
* **Logging:** `uber-go/zap`

---

## License

- **License:** This project is licensed under  [MIT](LICENSE)
- **Third-party Licenses:** The full license texts are available in the  [licenses/](licenses/)

