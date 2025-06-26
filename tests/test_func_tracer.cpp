#define CATCH_CONFIG_MAIN
#include "catch2/catch.hpp"
#include <string>
#include <set>
#include <mutex>

// Updated is_Relevant logic (returns true if function is relevant, i.e., NOT blacklisted)
bool is_Relevant(const std::string& func_name) {
    static const std::set<std::string> blacklist = {
        "main", "_init", "_start", ".plt.got"
    };
    if (func_name.length() >= 4 && func_name.compare(func_name.length() - 4, 4, "@plt") == 0) return false;
    if (func_name.length() >= 2 && func_name.compare(0, 2, "__") == 0) return false;
    return (blacklist.find(func_name) == blacklist.end());
}

// Deduplication logic for testing
static std::set<std::string> logged_functions;
static std::mutex log_mutex;

bool log_function_call_test(const char* img_name, const char* func_name)
{
    std::string key = std::string(img_name) + ":" + std::string(func_name);
    {
        std::lock_guard<std::mutex> guard(log_mutex);
        if (logged_functions.find(key) != logged_functions.end())
            return false; // Already logged, skip
        logged_functions.insert(key);
    }
    return true; // Logged for the first time
}

TEST_CASE("is_Relevant works as expected") {
    SECTION("PLT functions are not relevant") {
        REQUIRE_FALSE(is_Relevant("foo@plt"));
        REQUIRE_FALSE(is_Relevant("bar@plt"));
    }
    SECTION("Functions starting with __ are not relevant") {
        REQUIRE_FALSE(is_Relevant("__internal"));
        REQUIRE_FALSE(is_Relevant("__something"));
    }
    SECTION("Explicit blacklist") {
        REQUIRE_FALSE(is_Relevant("main"));
        REQUIRE_FALSE(is_Relevant("_init"));
        REQUIRE_FALSE(is_Relevant("_start"));
        REQUIRE_FALSE(is_Relevant(".plt.got"));
    }
    SECTION("Normal functions are relevant") {
        REQUIRE(is_Relevant("foo"));
        REQUIRE(is_Relevant("bar"));
        REQUIRE(is_Relevant("baz"));
    }
    SECTION("Short names are relevant") {
        REQUIRE(is_Relevant("a"));
        REQUIRE(is_Relevant("b@p"));
        REQUIRE(is_Relevant("_m"));
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
