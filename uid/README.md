# uid

Unique ID generation built on [google/uuid](https://github.com/google/uuid): compact 32-character hex IDs with optional entity prefixes.

- **Compact**: lowercase hex, no hyphens, always 32 characters.
- **Readable**: optional prefix such as `ord_` tells you the entity type at a glance in logs and databases.
- **Sortable option**: time-ordered IDs (UUID v7) that are friendly to database indexes.
- **Concurrency-safe**: every function can be called from multiple goroutines.

## Install

```bash
go get github.com/beyenpay/x/uid
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/beyenpay/x/uid"
)

// Define prefixes once instead of scattering string literals.
const PrefixOrder = "ord"

func main() {
	fmt.Println(uid.New())
	fmt.Println(uid.WithPrefix(PrefixOrder))
	fmt.Println(uid.NewSortable())
	fmt.Println(uid.WithPrefixSortable(PrefixOrder))
}
```

Output:

```
7daf3804ff224739b925b8a61b5cc550
ord_7daf3804ff224739b925b8a61b5cc550
019a3c5e7b2a7c4d8e1f0a2b3c4d5e6f
ord_019a3c5e7b2a7c4d8e1f0a2b3c4d5e6f
```

## API

| Function                  | Description                                          |
| ------------------------- | ---------------------------------------------------- |
| `New()`                   | Random ID (UUID v4), 32 hex characters               |
| `WithPrefix(p)`           | `New()` with `p_` in front; empty `p` returns `New()` |
| `NewSortable()`           | Time-ordered ID (UUID v7), 32 hex characters         |
| `WithPrefixSortable(p)`   | `NewSortable()` with `p_` in front                   |

## Which one to use

- Use `New` for tokens, request IDs and anything that should not leak creation time.
- Use `NewSortable` for database primary keys: later IDs sort after earlier ones, which avoids random index page splits. Note that it embeds a millisecond timestamp.

## Notes

- Prefixes should be short, lowercase and stable (`ord`, `pay`, `usr`). The separator is a single underscore.
- IDs are strings of fixed width, so a prefixed ID is `len(prefix) + 33` characters; size your columns accordingly.

## Layout

```
uid/
├── uid.go
├── uid_test.go
└── README.md
```
