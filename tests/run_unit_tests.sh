#!/bin/bash
set -e

echo "Running Python unit tests..."
export PIN_ROOT="${PIN_ROOT:-/var/coverage/pin}"
pushd tests || exit 1
python3 test_coverage_analyzer.py
sudo -E python3 test_wrap.py
echo "Building and running C++ unit tests..."
CXXFLAGS=$(pkg-config --cflags catch2 2>/dev/null || echo "")
LDFLAGS=$(pkg-config --libs catch2 2>/dev/null || echo "")
g++ -std=c++20 $CXXFLAGS test_func_tracer.cpp $LDFLAGS -o "test_func_tracer"
./test_func_tracer && rm -f test_func_tracer
popd 

echo "All tests passed."
