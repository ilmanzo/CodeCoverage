
# What is this ?
This repository is a code coverage tool that utilizes [Pin](https://www.intel.com/content/www/us/en/developer/articles/tool/pin-a-dynamic-binary-instrumentation-tool.html), a Dynamic Binary Instrumentation (DBI) engine published by Intel. It operates as a Pin plugin.

# Supported Platforms
- GNU/Linux

# Prerequisites:
- an x86_64 processor (doesn't work on other arch)
- make
- g++ 14 (15 doesn't work yet with Intel Pin)
- Catch2 library if you want to run the c++ testsuite : `zypper install Catch2-2-devel` 

# How to use (Quick Start)
First, clone this repository.
```
git clone git@github.com:ilmanzo/BinaryCoverage.git
```

Pin tool itself is required to run the code.
You can download the Pin tool from the Intel Pin website.

Scripts for downloading, building, and running examples are included in this repository.
Please start by running `build.sh.`

```
cd BinaryCoverage/
./build.sh
```

In `build.sh`, Pin 3.31 is downloaded and extracted to the same directory as this repository. Then, the source code in this repository is built.

To use the latest version of the pin, download it from the site as needed.

## Running the coverage tool
After downloading and building, please run 01_run_example.sh.

In `build.sh`, a C program included in example/ is executed as a coverage measurement target for this coverage tool.

After running, the coverage measurement results are output as a plain text file.

# Build and Execution Commands
## Build
To build this tool, please execute:

export PIN_ROOT=../pin-external-3.31-98869-gfa6f126a8-gcc-linux

```
make
```

The convention when building Pin tools is to specify the directory path of the PIN tool with PIN_ROOT.

# Execution
To run this tool, execute the following command:

```
$PIN_ROOT/pin -t ./obj-intel64/FuncTracer.so -- <target_module_path> <target_args...>
```

where <target_module_path> and <target_args...> are the path and any arguments for the target module you want to measure code coverage for.

# Note
This coverage tool uses DWARF debugging information to obtain line number information.
Pin 3.31 supports DWARF4 as debugging information. When building the application for which you want to measure coverage, please build it with the `-g` and `-gdwarf-4` options.

```
gcc -g -gdwarf-4 main.c
```


# Run unit tests

```
./tests/run_unit_tests.sh
```

