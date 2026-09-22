package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDisplaysCountsByDefault(t *testing.T) {
	output, err := runForTest([]string{"--bins=2"}, "1\n2\n3\n4\n")
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	normalized := normalizeWhitespace(output)
	if !strings.Contains(normalized, "Total count = 4") {
		t.Fatalf("missing total count in output: %q", output)
	}
	if !strings.Contains(normalized, "Min/Avg/Max = 1.00 / 2.50 / 4.00") {
		t.Fatalf("missing summary stats in output: %q", output)
	}
	if !strings.Contains(normalized, "[ 1.00, 2.50 ] 2") {
		t.Fatalf("missing first bin count in output: %q", output)
	}
	if !strings.Contains(normalized, "[ 2.50, 4.00 ] 2") {
		t.Fatalf("missing second bin count in output: %q", output)
	}
}

func TestRunDisplaysCountsAndPercentages(t *testing.T) {
	output, err := runForTest([]string{"--bins=2", "--percent"}, "1\n2\n3\n4\n")
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	normalized := normalizeWhitespace(output)
	if !strings.Contains(normalized, "[ 1.00, 2.50 ] 2 ( 50.00%)") {
		t.Fatalf("missing first bin count and percentage in output: %q", output)
	}
	if !strings.Contains(normalized, "[ 2.50, 4.00 ] 2 ( 50.00%)") {
		t.Fatalf("missing second bin count and percentage in output: %q", output)
	}
}

func TestRunDisplaysPercentiles(t *testing.T) {
	output, err := runForTest([]string{"--bins=5", "--percentiles=50,90,99"}, "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n")
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(output, "Percentiles = p50:5.00 p90:9.00 p99:10.00") {
		t.Fatalf("missing percentile summary in output: %q", output)
	}
}

func TestRunRejectsInvalidPercentiles(t *testing.T) {
	_, err := runForTest([]string{"--percentiles=0"}, "1\n2\n3\n")
	if err == nil {
		t.Fatal("expected error for invalid percentile")
	}
}

func TestRunHandlesSingleValueRanges(t *testing.T) {
	output, err := runForTest([]string{"--bins=3"}, "5\n5\n5\n")
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	normalized := normalizeWhitespace(output)
	if !strings.Contains(normalized, "Min/Avg/Max = 5.00 / 5.00 / 5.00") {
		t.Fatalf("missing summary stats in output: %q", output)
	}
	if !strings.Contains(normalized, "[ 5.00, 5.00 ] 3") {
		t.Fatalf("missing populated first bin in output: %q", output)
	}
}

func TestRunFiltersStatisticsByMinAndMax(t *testing.T) {
	output, err := runForTest(
		[]string{"--bins=2", "--min=2", "--max=4", "--percent", "--percentiles=50,100"},
		"1\n2\n3\n4\n5\n",
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	normalized := normalizeWhitespace(output)
	for _, want := range []string{
		"Total count = 3",
		"Min/Avg/Max = 2.00 / 3.00 / 4.00",
		"Percentiles = p50:3.00 p100:4.00",
		"[ 2.00, 3.00 ] 1 ( 33.33%)",
		"[ 3.00, 4.00 ] 2 ( 66.67%)",
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("output does not contain %q: %q", want, output)
		}
	}
}

func TestRunFiltersStatisticsWithOneSidedBounds(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "minimum", args: []string{"--min=3"}, want: "Total count = 2 Min/Avg/Max = 3.00 / 3.50 / 4.00"},
		{name: "maximum", args: []string{"--max=2"}, want: "Total count = 2 Min/Avg/Max = 1.00 / 1.50 / 2.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runForTest(tt.args, "1\n2\n3\n4\n")
			if err != nil {
				t.Fatalf("run returned error: %v", err)
			}
			if normalized := normalizeWhitespace(output); !strings.Contains(normalized, tt.want) {
				t.Fatalf("output does not contain %q: %q", tt.want, output)
			}
		})
	}
}

func TestRunHandlesNoValuesWithinBounds(t *testing.T) {
	output, err := runForTest([]string{"--min=10"}, "1\n2\n3\n")
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if output != "Total count = 0\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func runForTest(args []string, input string) (string, error) {
	var stdout bytes.Buffer
	err := run(args, strings.NewReader(input), &stdout, false)
	return stdout.String(), err
}

func normalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
