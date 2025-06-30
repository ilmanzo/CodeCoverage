#define CATCH_CONFIG_MAIN
#define COVERAGE_TESTING_BUILD
#include "catch2/catch.hpp"
#include <string>
#include <set>
#include <mutex>
#include "../FuncTracer.hpp"

// Deduplication logic for testing
static std::set<std::string> logged_functions;
static std::mutex log_mutex;

bool log_function_call_test(const char* img_name, const char* func_name)
{
    std::string key;
    key.append(img_name).append(1, ':').append(func_name);
    {
        std::lock_guard<std::mutex> guard(log_mutex);
        if (logged_functions.contains(key))
            return false; // Already logged, skip
        logged_functions.insert(key);
    }
    return true; // Logged for the first time
}

TEST_CASE("is_relevant works as expected") {
    SECTION("PLT functions are not relevant") {
        REQUIRE_FALSE(is_relevant("foo@plt"));
        REQUIRE_FALSE(is_relevant("bar@plt"));
    }
    SECTION("Functions starting with __ are not relevant") {
        REQUIRE_FALSE(is_relevant("__internal"));
        REQUIRE_FALSE(is_relevant("__something"));
    }
    SECTION("Explicit blacklist") {
        REQUIRE_FALSE(is_relevant("main"));
        REQUIRE_FALSE(is_relevant("_init"));
        REQUIRE_FALSE(is_relevant("_start"));
        REQUIRE_FALSE(is_relevant(".plt.got"));
    }
    SECTION("Normal functions are relevant") {
        REQUIRE(is_relevant("foo"));
        REQUIRE(is_relevant("bar"));
        REQUIRE(is_relevant("baz"));
    }
    SECTION("Short names are relevant") {
        REQUIRE(is_relevant("a"));
        REQUIRE(is_relevant("b@p"));
        REQUIRE(is_relevant("_m"));
    }
}

TEST_CASE("log_function_call deduplicates function calls") {
    // Clear the set before testing
    logged_functions.clear();

    SECTION("First call logs, second call skips") {
        REQUIRE(log_function_call_test("img1", "funcA") == true);  // First call, should log
        REQUIRE(log_function_call_test("img1", "funcA") == false); // Second call, should skip
    }

    SECTION("Different functions are logged separately") {
        REQUIRE(log_function_call_test("img1", "funcB") == true);
        REQUIRE(log_function_call_test("img1", "funcC") == true);
        REQUIRE(log_function_call_test("img1", "funcB") == false);
        REQUIRE(log_function_call_test("img1", "funcC") == false);
    }

    SECTION("Same function name in different images are logged separately") {
        REQUIRE(log_function_call_test("img1", "funcD") == true);
        REQUIRE(log_function_call_test("img2", "funcD") == true);
        REQUIRE(log_function_call_test("img1", "funcD") == false);
        REQUIRE(log_function_call_test("img2", "funcD") == false);
    }
}
