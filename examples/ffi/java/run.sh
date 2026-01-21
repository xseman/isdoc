#!/usr/bin/env bash
set -e
cd "$(dirname "${0}")" || exit 1

LIB_DIR="../../../bin"
if [ ! -f "${LIB_DIR}/libisdoc.so" ] && [ ! -f "${LIB_DIR}/libisdoc" ]; then
    echo "Error: libisdoc not found. Run make build-ffi first."
    exit 1
fi

if [ -f "${LIB_DIR}/libisdoc" ] && [ ! -f "${LIB_DIR}/libisdoc.so" ]; then
    ln -sf libisdoc "${LIB_DIR}/libisdoc.so"
fi

javac --enable-preview --release 21 Example.java
export LD_LIBRARY_PATH=${LIB_DIR}:$LD_LIBRARY_PATH
java --enable-preview --enable-native-access=ALL-UNNAMED Example
