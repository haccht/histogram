package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh/terminal"
)

type options struct {
	Bins int     `short:"b" long:"bins" description:"Number of bins in the histogram" default:"10"`
	Min  float64 `long:"min" description:"Minimum value in the histogram"`
	Max  float64 `long:"max" description:"Maximum value in the histogram"`
}

func run() error {
	var opts options
	args, err := flags.Parse(&opts)
	if err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
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
	if !terminal.IsTerminal(0) {
		readers = append(readers, os.Stdin)
	}
	if len(readers) == 0 {
		os.Exit(0)
	}

	var sum float64
	vals := make([]float64, 0, 1024*1024)

	scanner := bufio.NewScanner(io.MultiReader(readers...))
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			sum += val
			if opts.Min != 0 && opts.Min > val {
				continue
			}
			if opts.Max != 0 && opts.Max < val {
				continue
			}
			vals = append(vals, val)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(vals) == 0 {
		fmt.Printf("Total count = 0\n")
		return nil
	}

	min := opts.Min
	if min == 0 {
		min = slices.Min(vals)
	}

	max := opts.Max
	if max == 0 {
		max = slices.Max(vals)
	}

	w := (max - min) / float64(opts.Bins)

	var mcount int
	bins := make([]int, opts.Bins, opts.Bins)
	for _, val := range vals {
		var idx int
		switch {
		case min <= val && val < max:
			idx = int((val - min) / w)
		case val == max:
			idx = opts.Bins - 1
		}

		bins[idx]++
		if bins[idx] > mcount {
			mcount = bins[idx]
		}
	}

	fmt.Printf("Total count = %d\n", len(vals))
	fmt.Printf("Min/Avg/Max = %.2f / %.2f / %.2f\n\n", slices.Min(vals), sum/float64(len(vals)), slices.Max(vals))

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	for idx, count := range bins {
		bmin := fmt.Sprintf("%.2f", min+w*float64(idx))
		bmax := fmt.Sprintf("%.2f", min+w*float64(idx)+w)
		bar := "  " + strings.Repeat("|", 40*count/mcount)

		fmt.Fprintf(tw, "[\t%s,\t %s\t]\t%6d\t%s\n", bmin, bmax, count, bar)
	}
	tw.Flush()

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
