#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Running Python unit tests..."
PYTHONPATH="$SCRIPT_DIR/.." python3 "$SCRIPT_DIR/test_coverage_analyzer.py"

echo "Building and running C++ unit tests..."
CXXFLAGS=$(pkg-config --cflags catch2 2>/dev/null || echo "")
LDFLAGS=$(pkg-config --libs catch2 2>/dev/null || echo "")
g++ -std=c++11 $CXXFLAGS "$SCRIPT_DIR/test_func_tracer.cpp" $LDFLAGS -o "$SCRIPT_DIR/test_func_tracer"
"$SCRIPT_DIR/test_func_tracer"

echo "All tests passed."
