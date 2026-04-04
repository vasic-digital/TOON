# digital.vasic.toon

Generic reusable Go module for Token-Oriented Object Notation (TOON) encoding and decoding.

TOON is a compact, human-readable serialization format optimized for LLM token efficiency.

## Installation

```bash
go get digital.vasic.toon
```

## Usage

```go
import "digital.vasic.toon/pkg/toon"

// Encode
data, _ := toon.Marshal(myStruct)

// Decode
var result MyStruct
toon.Unmarshal(data, &result)

// Check content type
if toon.IsTOONContentType(r.Header.Get("Content-Type")) {
    // Handle TOON request
}

// Compare token efficiency
comp, _ := toon.Compare(myStruct)
fmt.Printf("Savings: %.1f%%\n", comp.Savings)
```
