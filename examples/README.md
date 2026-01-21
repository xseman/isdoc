# Examples

Usage examples for the isdoc library across different platforms and languages.

## Go Examples

- [Basic](basic/) - Simple ISDOC encoding and decoding
- [Advanced](advanced/) - Advanced features and validation

## FFI (Foreign Function Interface)

Use isdoc from other programming languages via C FFI:

- [Java](ffi/java/) - Using JNA
- [PHP](ffi/php/) - Using FFI extension
- [Python](ffi/python/) - Using ctypes
- [Swift](ffi/swift/) - Using C interoperability

See [FFI README](ffi/README.md) for setup instructions.

## CLI

Command-line usage examples:

- [CLI Examples](cli/) - Shell scripts for common operations

## Quick Start

**Go Examples:**

Build and run the basic example:

```bash
cd basic && go run main.go
```

**CLI:**

Build the CLI tool first:

```bash
cd .. && make build
```

Then use the examples:

```bash
cd examples/cli && ./validate.sh
```

**FFI:**

Build the library first:

```bash
cd ffi && ./build.sh
```

Then run language-specific examples (see individual directories).

## Related Documentation

- [Main README](../README.md)
- [API Documentation](../api.go)
