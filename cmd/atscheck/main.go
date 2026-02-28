package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"cut-the-bs/internal/atscheck"
)

type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("flag value cannot be empty")
	}
	*s = append(*s, trimmed)
	return nil
}

func main() {
	var pdfPath string
	var required stringSliceFlag
	var ordered stringSliceFlag
	var printText bool
	var strict bool

	flag.StringVar(&pdfPath, "pdf", "", "Path to PDF file to inspect")
	flag.Var(&required, "require", "Required substring (repeatable)")
	flag.Var(&ordered, "order", "Required order chain separated by '>' (repeatable)")
	flag.BoolVar(&printText, "print-text", false, "Print extracted text")
	flag.BoolVar(&strict, "strict", false, "Treat warnings as failures")
	flag.Parse()

	if strings.TrimSpace(pdfPath) == "" {
		fmt.Fprintln(os.Stderr, "atscheck: --pdf is required")
		flag.Usage()
		os.Exit(2)
	}

	orderedChecks, err := parseOrderedChecks(ordered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "atscheck: %v\n", err)
		os.Exit(2)
	}

	report, err := atscheck.AnalyzePDF(pdfPath, atscheck.Options{
		Required: []string(required),
		Ordered:  orderedChecks,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "atscheck: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ATS report for %s\n", pdfPath)
	if len(report.Errors) == 0 {
		fmt.Println("- errors: none")
	} else {
		fmt.Printf("- errors (%d):\n", len(report.Errors))
		for _, issue := range report.Errors {
			fmt.Printf("  - %s\n", issue)
		}
	}

	if len(report.Warnings) == 0 {
		fmt.Println("- warnings: none")
	} else {
		fmt.Printf("- warnings (%d):\n", len(report.Warnings))
		for _, issue := range report.Warnings {
			fmt.Printf("  - %s\n", issue)
		}
	}

	if printText {
		fmt.Println("\n--- Extracted Text ---")
		fmt.Println(report.Text)
		fmt.Println("--- End Extracted Text ---")
	}

	passed := report.Passed() && (!strict || len(report.Warnings) == 0)
	if passed {
		fmt.Println("PASS: ATS checks passed")
		return
	}

	if strict && report.Passed() && len(report.Warnings) > 0 {
		fmt.Println("FAIL: strict mode treats warnings as failures")
	} else {
		fmt.Println("FAIL: ATS checks failed")
	}
	os.Exit(1)
}

func parseOrderedChecks(values []string) ([][]string, error) {
	checks := make([][]string, 0, len(values))
	for _, raw := range values {
		parts := strings.Split(raw, ">")
		clean := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				clean = append(clean, trimmed)
			}
		}

		if len(clean) < 2 {
			return nil, fmt.Errorf("invalid --order value %q; use "+
				"marker1>marker2 (2+ markers)", raw)
		}

		checks = append(checks, clean)
	}

	return checks, nil
}
