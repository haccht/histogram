package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"golang.org/x/crypto/ssh/terminal"
)

type options struct {
	Bins  int  `short:"b" long:"bins" description:"Number of bins in the histogram" default:"10"`
	Chart bool `short:"c" long:"chart" description:"Draw the bar chart"`
}

func main() {
	var opts options
	args, err := flags.Parse(&opts)
	if err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	readers := make([]io.Reader, 0, len(args)+1)
	if !terminal.IsTerminal(0) {
		readers = append(readers, os.Stdin)
	}
	for _, arg := range args {
		f, err := os.Open(arg)
		if err != nil {
			log.Fatalf("failed to open file %s: %s", arg, err)
		}
		defer f.Close()
		readers = append(readers, f)
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
			vals = append(vals, val)
		}
	}
	if scanner.Err() != nil {
		log.Fatalln(scanner.Err())
	}

	sort.Float64s(vals)

	min := vals[0]
	max := vals[len(vals)-1]
	w := (max - min) / float64(opts.Bins)

	var mcount int
	bins := make([]int, opts.Bins, opts.Bins)

	for _, val := range vals {
		var idx int
		if val != max {
			idx = int((val - min) / w)
		} else {
			idx = opts.Bins - 1
		}

		bins[idx]++
		if bins[idx] > mcount {
			mcount = bins[idx]
		}
	}

	fmt.Printf("Total count = %d\n", len(vals))
	fmt.Printf("Min/Avg/Max = %.2f / %.2f / %.2f\n\n", min, w/float64(len(vals)), max)

	for idx, count := range bins {
		bmin := fmt.Sprintf("%8.2f", min+w*float64(idx))
		bmax := fmt.Sprintf("%8.2f", min+w*float64(idx)+w)
		bar := "  " + strings.Repeat("|", 40*count/mcount)

		fmt.Printf("[%s - %s]\t%8d", bmin, bmax, count)
		if opts.Chart {
			fmt.Print(bar)
		}
		fmt.Println("")
	}
}
