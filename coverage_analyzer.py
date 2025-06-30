#!/usr/bin/env python3

import argparse
import re
import os
from collections import defaultdict
import xml.etree.ElementTree as ET

# --- HTML Template ---
# Using a template string separates the HTML structure from the Python logic.
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Coverage Report for {image_name}</title>
    <style>
        body {{ font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; margin: 2em; background-color: #f9f9f9; color: #333; }}
        .container {{ max-width: 1200px; margin: auto; background-color: #fff; padding: 2em; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1); }}
        h1, h2 {{ color: #1a1a1a; border-bottom: 2px solid #eee; padding-bottom: 0.3em; }}
        .summary {{ background-color: #f4f4f4; padding: 1.5em; border-radius: 8px; margin-bottom: 2em; border: 1px solid #ddd; }}
        .summary p {{ margin: 0.5em 0; font-size: 1.1em; }}
        .summary .percentage {{ font-size: 1.8em; font-weight: bold; color: #0056b3; }}
        .progress-bar {{ background-color: #e9ecef; border-radius: 50px; overflow: hidden; height: 30px; margin-top: 1em; }}
        .progress-bar-inner {{ 
            background-color: #28a745; 
            height: 100%; 
            width: {coverage_percentage:.2f}%; 
            color: white; 
            text-align: center; 
            line-height: 30px;
            font-weight: bold;
            transition: width 0.5s ease-in-out;
        }}
        .function-list {{
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
            gap: 1em;
            list-style-type: none;
            padding: 0;
        }}
        .function-list li {{
            padding: 0.6em;
            border-radius: 5px;
            font-family: monospace;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            transition: transform 0.2s ease;
        }}
        .function-list li:hover {{
            transform: translateY(-2px);
            box-shadow: 0 2px 4px rgba(0,0,0,0.08);
        }}
        .called {{ background-color: #d4edda; color: #155724; border-left: 5px solid #28a745; }}
        .uncalled {{ background-color: #f8d7da; color: #721c24; border-left: 5px solid #dc3545; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>Coverage Report</h1>
        <h2>Image: {image_name}</h2>
            <div class="summary">
            <p><strong>Total Functions:</strong> {total_count}</p>
            <p><strong>Called Functions:</strong> {called_count}</p>
            <p><strong>Uncalled Functions:</strong> {uncalled_count}</p>
            <p class="percentage">Coverage: {coverage_percentage:.2f}%</p>
            <div class="progress-bar">
                <div class="progress-bar-inner">{coverage_percentage:.2f}%</div>
            </div>
        </div>
        <details>
        <summary><h2>Function Details</h2></summary>
        <p><strong>Legend: </strong><span class="called"> Called Function </span><span class="uncalled"> Uncalled Function </span></p>
        <p>List of functions found in the log file:</p>
        <ul class="function-list">
            {function_list_html}
        </ul>
        </details>
    </div>
</body>
</html>
"""

def analyze_logs(log_files):
    """
    Parses Intel Pin log files to generate a function call coverage report.
    """
    function_def_re = re.compile(r"\[Image:(?P<image>.*?)\] \[Function:(?P<function>.*?)\]")
    function_call_re = re.compile(r"\[Image:(?P<image>.*?)\] \[Called:(?P<function>.*?)\]")

    coverage_data = defaultdict(lambda: {"total_functions": set(), "called_functions": set()})

    print(f"--> Processing {len(log_files)} log file(s)...")

    for log_file in log_files:
        try:
            with open(log_file, 'r', encoding='utf-8', errors='ignore') as f:
                for line in f:
                    match_def = function_def_re.search(line)
                    if match_def:
                        image = match_def.group('image').strip()
                        function = match_def.group('function').strip()
                        if image and function:
                            coverage_data[image]['total_functions'].add(function)
                        continue

                    match_call = function_call_re.search(line)
                    if match_call:
                        image = match_call.group('image').strip()
                        function = match_call.group('function').strip()
                        if image and function:
                            coverage_data[image]['called_functions'].add(function)
        except FileNotFoundError:
            print(f"Warning: Log file not found: {log_file}")
        except Exception as e:
            print(f"An error occurred while reading {log_file}: {e}")

    print("--> Processing complete.\n")
    return coverage_data

def print_report(coverage_data):
    """
    Prints a formatted summary report from the analyzed coverage data to the console.
    """
    if not coverage_data:
        print("No data to report. Please check your log files.")
        return

    for image, data in sorted(coverage_data.items()):
        total_functions = data['total_functions']
        called_functions = data['called_functions']
        uncalled_functions = sorted(list(total_functions - called_functions))

        total_count = len(total_functions)
        called_count = len(called_functions)
        coverage_percentage = (called_count / total_count) * 100 if total_count > 0 else 0

        print("\n==================================================")
        print(f"Image: {image}")
        print("==================================================")
        print(f"  Functions Found:   {total_count}")
        print(f"  Functions Called:  {called_count}")
        print(f"  Coverage:          {coverage_percentage:.2f}%")
        print("--------------------------------------------------")

        if called_count > 0:
            print("  Called Functions:")
            for func in sorted(list(called_functions)):
                print(f"    - {func}")
        else:
            print("  No functions were called for this image.")
        
        if uncalled_functions:
            print("\n  Uncalled Functions:")
            for func in uncalled_functions:
                print(f"    - {func}")

    print("\n--- End of Console Report ---")

def generate_html_report(image_name, data, output_dir):
    """Generates an HTML coverage report for a single image using a template."""
    sanitized_name = re.sub(r'[^a-zA-Z0-9._-]', '_', os.path.basename(image_name))
    output_filename = os.path.join(output_dir, f"coverage_{sanitized_name}.html")

    total_functions = sorted(list(data['total_functions']))
    called_functions = data['called_functions']
    
    total_count = len(total_functions)
    called_count = len(called_functions)
    uncalled_count = total_count - called_count
    coverage_percentage = (called_count / total_count) * 100 if total_count > 0 else 0

    # Generate the HTML for the list of functions
    function_items = []
    for func in total_functions:
        status_class = "called" if func in called_functions else "uncalled"
        function_items.append(f'<li class="{status_class}" title="{func}">{func}</li>')
    function_list_html = "\n".join(function_items)

    # Populate the main HTML template
    html_content = HTML_TEMPLATE.format(
        image_name=sanitized_name,
        total_count=total_count,
        called_count=called_count,
        uncalled_count=uncalled_count,
        coverage_percentage=coverage_percentage,
        function_list_html=function_list_html
    )

    try:
        with open(output_filename, 'w', encoding='utf-8') as f:
            f.write(html_content)
        print(f"    - HTML report saved to {output_filename}")
    except Exception as e:
        print(f"Error writing HTML file {output_filename}: {e}")


def generate_xunit_report(image_name, data, output_dir):
    """Generates an XUnit XML coverage report for a single image."""
    sanitized_name = re.sub(r'[^a-zA-Z0-9._-]', '_', os.path.basename(image_name))
    output_filename = os.path.join(output_dir, f"coverage_{sanitized_name}.xml")
    
    total_functions = sorted(list(data['total_functions']))
    called_functions = data['called_functions']
    total_count = len(total_functions)
    skipped_count = total_count - len(called_functions)

    testsuites = ET.Element("testsuites")
    testsuite = ET.SubElement(
        testsuites, "testsuite",
        {
            "errors": "0",
            "failures": "0",
            "name": f"binary_coverage_{str(sanitized_name)}",
            "skipped": str(skipped_count),
            "tests": str(total_count),
        }
    )

    # Single testcase with all function information
    testcase = ET.SubElement(
        testsuite, "testcase",
        {
            "classname": f"binary_coverage_{sanitized_name}",
            "name": "Result",
        }
    )
    
    called_list = sorted(list(called_functions))
    uncalled_list = sorted(list(data['total_functions'] - called_functions))
    
    # The message which is always visible
    summary_parts = []
    summary_parts.append(f"Coverage Summary for {sanitized_name}")
    summary_parts.append(f"Total Functions: {total_count}")
    summary_parts.append(f"Called Functions: {len(called_functions)}")
    summary_parts.append(f"Uncalled Functions: {skipped_count}")
    summary_parts.append(f"Coverage: {(len(called_functions) / total_count * 100):.2f}%" if total_count > 0 else "Coverage: 0.00%")
    summary_message = " | ".join(summary_parts)
    
    # The detailed lists of functions
    function_details = []
    
    if called_list:
        function_details.append("CALLED FUNCTIONS:")
        for func in called_list:
            function_details.append(f"  ✓ {func}")
        function_details.append("")
    
    if uncalled_list:
        function_details.append("UNCALLED FUNCTIONS:")
        for func in uncalled_list:
            function_details.append(f"  ✗ {func}")
    
    passed_element = ET.SubElement(testcase, "passed")
    passed_element.set("message", summary_message)
    passed_element.text = "\n".join(function_details)

    tree = ET.ElementTree(testsuites)
    try:
        tree.write(output_filename, encoding="utf-8", xml_declaration=True)
        print(f"    - XUnit report saved to {output_filename}")
    except Exception as e:
            print(f"Error writing XUnit file {output_filename}: {e}")

def main():
    """Main function to parse arguments and run the analysis."""
    parser = argparse.ArgumentParser(
        description="Analyze Intel Pin logs for function call coverage.",
        formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument(
        'log_files',
        nargs='+',
        help='One or more log files to analyze.'
    )
    parser.add_argument(
        '--html-output',
        metavar='<directory>',
        type=str,
        help='Directory to save per-image HTML coverage reports.'
    )
    parser.add_argument(
        '--xunit-output',
        metavar='<directory>',
        type=str,
        help='Directory to save per-image XUnit XML coverage reports.'
    )
    args = parser.parse_args()

    coverage_data = analyze_logs(args.log_files)
    print_report(coverage_data)

    if args.html_output:
        output_dir = args.html_output
        try:
            os.makedirs(output_dir, exist_ok=True)
            print(f"\n--> Generating HTML reports in '{output_dir}'...")
            for image, data in sorted(coverage_data.items()):
                generate_html_report(image, data, output_dir)
            print("--> HTML generation complete.")
        except OSError as e:
            print(f"Error: Could not create HTML output directory '{output_dir}': {e}")

    if args.xunit_output:
        output_dir = args.xunit_output
        try:
            os.makedirs(output_dir, exist_ok=True)
            print(f"\n--> Generating XUnit reports in '{output_dir}'...")
            for image, data in sorted(coverage_data.items()):
                generate_xunit_report(image, data, output_dir)
            print("--> XUnit generation complete.")
        except OSError as e:
            print(f"Error: Could not create XUnit output directory '{output_dir}': {e}")

if __name__ == '__main__':
    main()