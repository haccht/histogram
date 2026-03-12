package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh/terminal"
)

type options struct {
	Bins        int            `short:"b" long:"bins" description:"Number of bins in the histogram" default:"10"`
	Min         *float64       `long:"min" description:"Minimum value in the histogram"`
	Max         *float64       `long:"max" description:"Maximum value in the histogram"`
	Percent     bool           `long:"percent" description:"Display bin values as percentages"`
	Percentiles percentileList `long:"percentiles" description:"Comma-separated percentiles to display (e.g. 50,90,99)"`
}

type percentileList []float64

func (p *percentileList) UnmarshalFlag(value string) error {
	for _, field := range strings.Split(value, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			return fmt.Errorf("percentiles must not contain empty values")
		}

		percentile, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return fmt.Errorf("invalid percentile %q: %w", field, err)
		}
		if percentile <= 0 || percentile > 100 {
			return fmt.Errorf("percentiles must be between 0 and 100: %s", field)
		}

		*p = append(*p, percentile)
	}
	return nil
}

func parseOptions(args []string) (options, []string, error) {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	remaining, err := parser.ParseArgs(args)
	if err != nil {
		return options{}, nil, err
	}
	if opts.Bins <= 0 {
		return options{}, nil, fmt.Errorf("--bins must be greater than 0")
	}
	if opts.Min != nil && opts.Max != nil && *opts.Min > *opts.Max {
		return options{}, nil, fmt.Errorf("--min must be less than or equal to --max")
	}

	return opts, remaining, nil
}

func run(args []string, stdin io.Reader, stdout io.Writer, stdinIsTerminal bool) error {
	opts, args, err := parseOptions(args)
	if err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			return nil
		}
		return err
	}

	readers := make([]io.Reader, 0, len(args)+1)
	for _, arg := range args {
		f, err := os.Open(arg)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %s", arg, err)
		}
		defer f.Close()
		readers = append(readers, f)
	}
	if !stdinIsTerminal {
		readers = append(readers, stdin)
	}
	if len(readers) == 0 {
		return nil
	}

	return renderHistogram(opts, readers, stdout)
}

func renderHistogram(opts options, readers []io.Reader, stdout io.Writer) error {
	var sum float64
	vals := make([]float64, 0, 1024*1024)

	scanner := bufio.NewScanner(io.MultiReader(readers...))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			sum += val
			vals = append(vals, val)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(vals) == 0 {
		fmt.Fprintln(stdout, "Total count = 0")
		return nil
	}

	slices.Sort(vals)

	inputMin := vals[0]
	inputMax := vals[len(vals)-1]

	minVal := inputMin
	if opts.Min != nil {
		minVal = *opts.Min
	}
	maxVal := inputMax
	if opts.Max != nil {
		maxVal = *opts.Max
	}

	bins, mcount := buildBins(vals, opts.Bins, minVal, maxVal)

	fmt.Fprintf(stdout, "Total count = %d\n", len(vals))
	fmt.Fprintf(stdout, "Min/Avg/Max = %.2f / %.2f / %.2f\n", inputMin, sum/float64(len(vals)), inputMax)
	if len(opts.Percentiles) > 0 {
		fmt.Fprintf(stdout, "Percentiles = %s\n", formatPercentiles(vals, opts.Percentiles))
	}
	fmt.Fprintln(stdout)

	w := 0.0
	if opts.Bins > 0 {
		w = (maxVal - minVal) / float64(opts.Bins)
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	for idx, count := range bins {
		bmin := fmt.Sprintf("%.2f", minVal+w*float64(idx))
		bmax := fmt.Sprintf("%.2f", minVal+w*float64(idx)+w)
		bar := ""
		if mcount > 0 {
			bar = "  " + strings.Repeat("|", 40*count/mcount)
		}

		fmt.Fprintf(tw, "[\t%s,\t %s\t]\t%s\t%s\n", bmin, bmax, formatBinValue(count, len(vals), opts.Percent), bar)
	}
	return tw.Flush()
}

func buildBins(vals []float64, binCount int, minVal, maxVal float64) ([]int, int) {
	bins := make([]int, binCount)
	if binCount == 0 {
		return bins, 0
	}

	if minVal == maxVal {
		for _, val := range vals {
			if val == minVal {
				bins[0]++
			}
		}
		return bins, bins[0]
	}

	w := (maxVal - minVal) / float64(binCount)
	mcount := 0
	for _, val := range vals {
		switch {
		case val < minVal || val > maxVal:
			continue
		case val == maxVal:
			bins[binCount-1]++
			if bins[binCount-1] > mcount {
				mcount = bins[binCount-1]
			}
		default:
			idx := int((val - minVal) / w)
			if idx < 0 || idx >= binCount {
				continue
			}
			bins[idx]++
			if bins[idx] > mcount {
				mcount = bins[idx]
			}
		}
	}

	return bins, mcount
}

func formatBinValue(count, total int, percent bool) string {
	if percent {
		if total == 0 {
			return "  0.00%"
		}
		return fmt.Sprintf("%6.2f%%", 100*float64(count)/float64(total))
	}
	return fmt.Sprintf("%6d", count)
}

func formatPercentiles(vals []float64, percentiles percentileList) string {
	parts := make([]string, 0, len(percentiles))
	for _, p := range percentiles {
		label := strconv.FormatFloat(p, 'f', -1, 64)
		parts = append(parts, fmt.Sprintf("p%s:%.2f", label, percentile(vals, p)))
	}
	return strings.Join(parts, " ")
}

func percentile(vals []float64, p float64) float64 {
	if len(vals) == 0 {
		return 0
	}

	rank := int(math.Ceil(p / 100 * float64(len(vals))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(vals) {
		rank = len(vals)
	}

	return vals[rank-1]
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, terminal.IsTerminal(0)); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
