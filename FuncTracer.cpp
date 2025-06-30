/* FuncTracer.cpp */
#include "FuncTracer.hpp"

// Global set and mutex to track logged functions
static std::set<std::string> logged_functions;
static std::mutex log_mutex;

VOID log_function_call(const char* img_name, const char* func_name)
{
    // Check if the function is already logged to avoid duplicate logging
    std::string key;
    key.append(img_name).append(1, ':').append(func_name);
    {
        std::lock_guard<std::mutex> guard(log_mutex);
        if (logged_functions.contains(key))
            return; // Already logged, skip
        logged_functions.insert(key);
    }

    std::stringstream ss;
    PIN_LockClient();
    pid_t pid = PIN_GetPid();
    PIN_UnlockClient();
    ss << "[PID:" << pid << "] [Image:" << img_name << "] [Called:" << func_name << "]\n";
    LOG(ss.str());
}

// Determine if function name is relevant to us and if it will be logged
bool func_is_relevant(const std::string_view &func_name)
{
    static const std::set<std::string_view> blacklist = {
        "main", "_init", "_start", ".plt.got"
    };
    if (blacklist.contains(func_name))
        return false;

    // Ignore PLT functions and internal functions (usually prefixed with __)
    if (func_name.starts_with("@plt") || func_name.starts_with("__"))
        return false;

    return true;
}

// Pin calls this function for every image loaded into the process's address space.
// An image is either an executable or a shared library.
VOID image_load(IMG img, VOID *v)
{
    // We iterate through all the sections of the image.
    for (SEC sec = IMG_SecHead(img); SEC_Valid(sec); sec = SEC_Next(sec))
    {
        const std::string &image_name = IMG_Name(img);

        LOG("[Image:" + image_name + "] [Section:" + SEC_Name(sec) + "]\n");
        // We iterate through all the routines (functions) in the image.
        if (SEC_Type(sec) != SEC_TYPE_EXEC)
            continue; // Only instrument executable sections
        for (RTN rtn = SEC_RtnHead(sec); RTN_Valid(rtn); rtn = RTN_Next(rtn))
        {
            RTN_Open(rtn);
            const std::string &rtn_name = RTN_Name(rtn);
            if (func_is_relevant(rtn_name)) // Check if the function is relevant for our analysis
            {
                std::stringstream ss;
                // We log the image name and function name so we can see which function is being instrumented.
                ss << "[Image:" << image_name << "] [Function:" << RTN_Name(rtn) << "]\n";
                LOG(ss.str());
                // For each routine, we insert a call to our analysis function `log_function_call`.
                RTN_InsertCall(rtn, IPOINT_BEFORE, (AFUNPTR)log_function_call,
                               IARG_PTR, image_name.c_str(),
                               IARG_PTR, rtn_name.c_str(),
                               IARG_END);
            }
            RTN_Close(rtn);
        }
    }
}

// Pin calls this function when the application is about to fork a new process.
// Returning TRUE tells Pin to follow and instrument the child process.
BOOL follow_child_process(CHILD_PROCESS childProcess, VOID *v)
{
    // LOG( "New PID: " + decstr(PIN_GetPid()) + "\n");
    // TraceFile << "[PID: " << PIN_GetPid() << "] Forking a new process..." << std::endl;
    // TraceFile.flush();
    return TRUE; // Follow the child
}

// Pintool entry point
int main(int argc, char *argv[])
{

    // Initialize PIN symbols. This is required for routine-level instrumentation.
    PIN_InitSymbols();

    // Initialize PIN. This must be the first function called.
    if (PIN_Init(argc, argv))
    {
        std::cerr << "PIN_Init failed" << std::endl;
        return 1;
    }
    // Register the function to be called for every loaded image.
    IMG_AddInstrumentFunction(image_load, 0);

    // TODO: check if childs are automatically followed or not
    PIN_AddFollowChildProcessFunction(follow_child_process, 0);

    // Start the program, never returns
    PIN_StartProgram();

    return 0;
}
