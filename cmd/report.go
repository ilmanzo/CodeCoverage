package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const detailedHTMLTemplateStr = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Coverage Report for {{.ImageName}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 2em; background: #f9f9f9; color: #333; }
        .container { max-width: 900px; margin: auto; background: #fff; padding: 2em; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);}
        .summary { background: #f4f4f4; padding: 1.5em; border-radius: 8px; margin-bottom: 2em; border: 1px solid #ddd;}
        .summary .percentage { font-size: 1.8em; font-weight: bold; color: #0056b3;}
        .progress-bar { background: #e9ecef; border-radius: 50px; overflow: hidden; height: 30px; margin-top: 1em;}
        .progress-bar-inner { background: #28a745; height: 100%; color: white, text-align: center; line-height: 30px; font-weight: bold; transition: width 0.5s;}
        .function-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1em; list-style-type: none; padding: 0;}
        .function-list li { padding: 0.6em; border-radius: 5px; font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; transition: transform 0.2s;}
        .function-list li:hover { transform: translateY(-2px); box-shadow: 0 2px 4px rgba(0,0,0,0.08);}
        .called { background: #d4edda; color: #155724; border-left: 5px solid #28a745;}
        .uncalled { background: #f8d7da; color: #721c24; border-left: 5px solid #dc3545;}
    </style>
</head>
<body>
<div class="container">
    <h1>Coverage Report</h1>
    <h2>Image: {{.ImageName}}</h2>
    <div class="summary">
        <p><strong>Total Functions:</strong> {{.TotalCount}}</p>
        <p><strong>Called Functions:</strong> {{.CalledCount}}</p>
        <p><strong>Uncalled Functions:</strong> {{.UncalledCount}}</p>
        <p class="percentage">Coverage: {{printf "%.2f" .CoveragePercentage}}%</p>
        <div class="progress-bar">
            <div class="progress-bar-inner" style="width: {{.CoveragePercentage}}%">{{printf "%.2f" .CoveragePercentage}}%</div>
        </div>
    </div>
    <details>
    <summary><h2>Function Details</h2></summary>
    <p><strong>Legend: </strong><span class="called"> Called Function </span><span class="uncalled"> Uncalled Function </span></p>
    <ul class="function-list">
        {{range .Functions}}
        <li class="{{.Status}}" title="{{.Name}}">{{.Name}}</li>
        {{end}}
    </ul>
    </details>
</div>
</body>
</html>`

type CoverageData struct {
	TotalFunctions  map[string]struct{}
	CalledFunctions map[string]struct{}
}

type FunctionEntry struct {
	Name   string
	Status string // "called" or "uncalled"
}

type HTMLReportData struct {
	ImageName          string
	TotalCount         int
	CalledCount        int
	UncalledCount      int
	CoveragePercentage float64
	Functions          []FunctionEntry
}

// --- Coverage Analysis ---

var (
	functionDefRe  = regexp.MustCompile(`\[Image:(.*?)\] \[Function:(.*?)\]`)
	functionCallRe = regexp.MustCompile(`\[Image:(.*?)\] \[Called:(.*?)\]`)
)

func analyzeLogs(logFiles []string) (map[string]*CoverageData, error) {
	coverage := make(map[string]*CoverageData)
	for _, logFile := range logFiles {
		f, err := os.Open(logFile)
		if err != nil {
			return nil, fmt.Errorf("could not open log file %s: %w", logFile, err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if m := functionDefRe.FindStringSubmatch(line); m != nil {
				image, function := strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
				if image == "" || function == "" {
					continue
				}
				if _, ok := coverage[image]; !ok {
					coverage[image] = &CoverageData{make(map[string]struct{}), make(map[string]struct{})}
				}
				coverage[image].TotalFunctions[function] = struct{}{}
			} else if m := functionCallRe.FindStringSubmatch(line); m != nil {
				image, function := strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
				if image == "" || function == "" {
					continue
				}
				if _, ok := coverage[image]; !ok {
					coverage[image] = &CoverageData{make(map[string]struct{}), make(map[string]struct{})}
				}
				coverage[image].CalledFunctions[function] = struct{}{}
			}
		}
		f.Close()
	}
	return coverage, nil
}

func printTxtReport(coverage map[string]*CoverageData) {
	for image, data := range coverage {
		total := len(data.TotalFunctions)
		called := len(data.CalledFunctions)
		uncalled := total - called
		coveragePct := 0.0
		if total > 0 {
			coveragePct = float64(called) / float64(total) * 100
		}
		fmt.Printf("\n==================================================\n")
		fmt.Printf("Image: %s\n", image)
		fmt.Printf("==================================================\n")
		fmt.Printf("  Functions Found:   %d\n", total)
		fmt.Printf("  Functions Called:  %d\n", called)
		fmt.Printf("  Coverage:          %.2f%%\n", coveragePct)
		fmt.Printf("--------------------------------------------------\n")
		if called > 0 {
			fmt.Println("  Called Functions:")
			for fn := range data.CalledFunctions {
				fmt.Printf("    - %s\n", fn)
			}
		} else {
			fmt.Println("  No functions were called for this image.")
		}
		if uncalled > 0 {
			fmt.Println("\n  Uncalled Functions:")
			for fn := range data.TotalFunctions {
				if _, ok := data.CalledFunctions[fn]; !ok {
					fmt.Printf("    - %s\n", fn)
				}
			}
		}
	}
	fmt.Println("\n--- End of Console Report ---")
}

func generateHTMLReport(image string, data *CoverageData, outputDir string) error {
	totalFns := make([]string, 0, len(data.TotalFunctions))
	for fn := range data.TotalFunctions {
		totalFns = append(totalFns, fn)
	}
	calledFns := data.CalledFunctions
	totalCount := len(totalFns)
	calledCount := len(calledFns)
	uncalledCount := totalCount - calledCount
	coveragePct := 0.0
	if totalCount > 0 {
		coveragePct = float64(calledCount) / float64(totalCount) * 100
	}
	functions := make([]FunctionEntry, 0, totalCount)
	for _, fn := range totalFns {
		status := "uncalled"
		if _, ok := calledFns[fn]; ok {
			status = "called"
		}
		functions = append(functions, FunctionEntry{Name: fn, Status: status})
	}
	reportData := HTMLReportData{
		ImageName:          image,
		TotalCount:         totalCount,
		CalledCount:        calledCount,
		UncalledCount:      uncalledCount,
		CoveragePercentage: coveragePct,
		Functions:          functions,
	}
	tmpl, err := template.New("report").Parse(detailedHTMLTemplateStr)
	if err != nil {
		return err
	}
	safeName := regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(filepath.Base(image), "_")
	outfile := filepath.Join(outputDir, fmt.Sprintf("coverage_%s.html", safeName))
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, reportData)
}

// --- XUnit XML Report ---

type TestSuites struct {
	XMLName   xml.Name    `xml:"testsuites"`
	TestSuite []TestSuite `xml:"testsuite"`
}
type TestSuite struct {
	Errors   int        `xml:"errors,attr"`
	Failures int        `xml:"failures,attr"`
	Name     string     `xml:"name,attr"`
	Skipped  int        `xml:"skipped,attr"`
	Tests    int        `xml:"tests,attr"`
	TestCase []TestCase `xml:"testcase"`
}
type TestCase struct {
	ClassName string  `xml:"classname,attr"`
	Name      string  `xml:"name,attr"`
	Passed    *Passed `xml:"passed"`
}
type Passed struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

func generateXUnitReport(image string, data *CoverageData, outputDir string) error {
	totalFns := make([]string, 0, len(data.TotalFunctions))
	for fn := range data.TotalFunctions {
		totalFns = append(totalFns, fn)
	}
	calledFns := data.CalledFunctions
	totalCount := len(totalFns)
	skippedCount := totalCount - len(calledFns)
	calledList := make([]string, 0, len(calledFns))
	uncalledList := make([]string, 0, skippedCount)
	for fn := range data.TotalFunctions {
		if _, ok := calledFns[fn]; ok {
			calledList = append(calledList, fn)
		} else {
			uncalledList = append(uncalledList, fn)
		}
	}
	safeName := regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(filepath.Base(image), "_")
	outfile := filepath.Join(outputDir, fmt.Sprintf("coverage_%s.xml", safeName))

	summary := fmt.Sprintf("Coverage Summary for %s | Total Functions: %d | Called Functions: %d | Uncalled Functions: %d | Coverage: %.2f%%",
		safeName, totalCount, len(calledFns), skippedCount, float64(len(calledFns))/float64(totalCount)*100)
	var details strings.Builder
	if len(calledList) > 0 {
		details.WriteString("CALLED FUNCTIONS:\n")
		for _, fn := range calledList {
			details.WriteString(fmt.Sprintf("  ✓ %s\n", fn))
		}
		details.WriteString("\n")
	}
	if len(uncalledList) > 0 {
		details.WriteString("UNCALLED FUNCTIONS:\n")
		for _, fn := range uncalledList {
			details.WriteString(fmt.Sprintf("  ✗ %s\n", fn))
		}
	}
	ts := TestSuites{
		TestSuite: []TestSuite{
			{
				Errors:   0,
				Failures: 0,
				Name:     "binary_coverage_" + safeName,
				Skipped:  skippedCount,
				Tests:    totalCount,
				TestCase: []TestCase{
					{
						ClassName: "binary_coverage_" + safeName,
						Name:      "Result",
						Passed: &Passed{
							Message: summary,
							Text:    details.String(),
						},
					},
				},
			},
		},
	}
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	return enc.Encode(ts)
}

func generateAggregateHTMLReport(coverage map[string]*CoverageData, outputDir string) error {
	type Row struct {
		ImageName      string
		TotalCount     int
		CalledCount    int
		CoveragePct    float64
	}
	rows := []Row{}
	for image, data := range coverage {
		total := len(data.TotalFunctions)
		called := len(data.CalledFunctions)
		coveragePct := 0.0
		if total > 0 {
			coveragePct = float64(called) / float64(total) * 100
		}
		rows = append(rows, Row{
			ImageName:   image,
			TotalCount:  total,
			CalledCount: called,
			CoveragePct: coveragePct,
		})
	}

	const tpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Aggregate Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 2em; background: #f9f9f9; color: #333; }
        .container { max-width: 900px; margin: auto; background: #fff; padding: 2em; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);}
        table { width: 100%; border-collapse: collapse; margin-top: 2em;}
        th, td { padding: 0.7em 1em; border-bottom: 1px solid #ddd; text-align: left;}
        th { background: #f4f4f4; }
        tr:hover { background: #f1f7ff; }
        .bar { height: 18px; background: #e9ecef; border-radius: 9px; overflow: hidden; }
        .bar-inner { background: #28a745; height: 100%; color: white; text-align: center; font-size: 0.9em; font-weight: bold; }
    </style>
</head>
<body>
<div class="container">
    <h1>Aggregate Coverage Report</h1>
    <table>
        <tr>
            <th>Image</th>
            <th>Total Functions</th>
            <th>Called Functions</th>
            <th>Coverage</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.ImageName}}</td>
            <td>{{.TotalCount}}</td>
            <td>{{.CalledCount}}</td>
            <td>
                <div class="bar">
                    <div class="bar-inner" style="width: {{printf "%.2f" .CoveragePct}}%">{{printf "%.2f" .CoveragePct}}%</div>
                </div>
            </td>
        </tr>
        {{end}}
    </table>
</div>
</body>
</html>`

	tmpl, err := template.New("aggregate").Parse(tpl)
	if err != nil {
		return err
	}
	outfile := filepath.Join(outputDir, "aggregate_coverage.html")
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, rows)
}
