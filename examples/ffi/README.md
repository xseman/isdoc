# FFI Examples

Foreign Function Interface examples for using ISDOC library from different languages.

## Build

```bash
make build-ffi
```

Or build and test all examples:

```bash
bash build.sh
```

## Languages

- **Python** - Uses ctypes
- **PHP** - Uses FFI extension (PHP 7.4+)
- **Swift** - Uses C interop
- **Java** - Uses Foreign Function & Memory API (JDK 21+)

## API

```c
char* isdoc_parse(char* xmlData);
char* isdoc_validate(char* xmlData);
void isdoc_free(char* ptr);
```

All returned strings must be freed with `isdoc_free()`.

```bash
# Linux
export LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH

# macOS
export DYLD_LIBRARY_PATH=/path/to/lib:$DYLD_LIBRARY_PATH

# Windows
set PATH=C:\path\to\lib;%PATH%
```

### Symbol not found

Ensure you're using the correct library version. Check with:

```bash
# Linux/macOS
nm -D libisdoc.so | grep isdoc_

# Windows (using dumpbin)
dumpbin /exports libisdoc.dll
```
