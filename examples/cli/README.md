# CLI Examples

Command-line usage examples for the isdoc tool.

## Example Files

- [sample.isdoc](../../testdata/fixtures/sample.isdoc) - Sample ISDOC invoice
- [validate.sh](validate.sh) - Validate an ISDOC file
- [encode-decode.sh](encode-decode.sh) - Encode and decode example

## Usage

First, build the CLI binary:

```bash
cd ../.. && make build
```

### Validate an ISDOC file

```bash
./validate.sh
```

Or directly:

```bash
../../bin/isdoc validate ../../testdata/fixtures/sample.isdoc
```

### Encode and Decode

```bash
./encode-decode.sh
```

## Available Commands

The `isdoc` CLI provides the following commands:

- `validate <file>` - Validate an ISDOC XML file
- `encode <file>` - Encode ISDOC data to XML
- `decode <file>` - Decode ISDOC XML to JSON

Run `isdoc --help` for more information.
