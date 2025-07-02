#!/bin/bash
set -e

echo "Running Go unit tests..."
pushd cmd
go test -v ./...
popd
export PIN_ROOT="${PIN_ROOT:-/var/coverage/pin}"
pushd tests || exit 1
echo "Building and running C++ unit tests..."
CXXFLAGS=$(pkg-config --cflags catch2 2>/dev/null || echo "")
LDFLAGS=$(pkg-config --libs catch2 2>/dev/null || echo "")
g++ -std=c++20 $CXXFLAGS test_func_tracer.cpp $LDFLAGS -o "test_func_tracer"
./test_func_tracer && rm -f test_func_tracer
popd 

echo "All tests passed."
