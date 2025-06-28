#!/bin/bash
set -e


echo "Running Python unit tests..."
export PIN_ROOT="${PIN_ROOT:-/var/coverage/pin}"
pushd tests || exit 1
python3 test_coverage_analyzer.py
sudo -E python3 test_wrap.py
popd 

echo "Building and running C++ unit tests..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CXXFLAGS=$(pkg-config --cflags catch2 2>/dev/null || echo "")
LDFLAGS=$(pkg-config --libs catch2 2>/dev/null || echo "")
g++ -std=c++11 $CXXFLAGS "$SCRIPT_DIR/test_func_tracer.cpp" $LDFLAGS -o "$SCRIPT_DIR/test_func_tracer"
"$SCRIPT_DIR/test_func_tracer"

echo "All tests passed."
