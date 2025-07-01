/* FuncTracer.cpp */

// We do not want to include PIN in testing build
#ifndef COVERAGE_TESTING_BUILD
#include "pin.H"
#endif

#include <iostream>
#include <fstream>
#include <string>
#include <set>
#include <mutex>
#include <unistd.h> // For getpid()

VOID log_function_call(const char* img_name, const char* func_name);

// Determine if function name is relevant to us and if it will be logged
bool func_is_relevant(const std::string_view &func_name);

// Pin calls this function for every image loaded into the process's address space.
// An image is either an executable or a shared library.
VOID image_load(IMG img, VOID *v);

// Pin calls this function when the application is about to fork a new process.
// Returning TRUE tells Pin to follow and instrument the child process.
BOOL follow_child_process(CHILD_PROCESS childProcess, VOID *v);

// Pintool entry point
int main(int argc, char *argv[]);
