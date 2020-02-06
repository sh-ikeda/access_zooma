package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func is_valid_query(query string, ignoring_queries []string) bool {
	// `query` must not be a one-character string
	if len(query)==1 {
		return false
	}

	// `query` must not match any words in `ignoring_queries`
	for _, ignore := range ignoring_queries {
		if query == ignore {
			return false
		}
	}

	// `query` must not be a two-or-less digits number
	r1 := regexp.MustCompile(`^[0-9][0-9]?$`)
	if r1.MatchString(query) {
		return false
	}

	return true
}

func get_zooma_json(query string, split_line []string, isFin chan interface{}, wg *sync.WaitGroup) {
	const base string = "https://www.ebi.ac.uk/spot/zooma/v2/api/services/annotate?propertyValue="

	defer wg.Done()
	resp, err := http.Get(base + string(query))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bytes, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	str := string(bytes)
	if str != "[]" && str[0:1] == "[" {
		var decode_data interface{}
		if err := json.Unmarshal(bytes, &decode_data); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, query)
			fmt.Fprintln(os.Stderr, string(bytes))
			if err, ok := err.(*json.SyntaxError); ok {
				fmt.Fprintln(os.Stderr, string(bytes[err.Offset-15:err.Offset+15]))
			}
			os.Exit(1)
		}
		for _, data := range decode_data.([]interface{}) {
			d := data.(map[string]interface{})
			ap := d["annotatedProperty"].(map[string]interface{})
			var prtype string
			if ap["propertyType"] == nil {
				prtype = ""
			} else {
				prtype = ap["propertyType"].(string)
			}
			prvalue := ap["propertyValue"].(string)
			conf := d["confidence"].(string)
			t := d["semanticTags"].([]interface{})
			term := t[0].(string)

			// Output to stdout must be done ONLY once per routine in concurrent processes. Multiple outputs within a routine can be interrputed by another routine.
			var unsplit_value string
			if len(split_line) == 4 {
				unsplit_value = split_line[3]
			} else {
				unsplit_value = ""
			}

			fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				split_line[0], split_line[1],
				split_line[2], prtype, prvalue, conf, term, unsplit_value)
				//query, prtype, prvalue, conf, term, unsplit_value)
			// if len(split_line) == 4 {
			// 	fmt.Fprintln(os.Stdout, split_line[3])
			// } else {
			// 	fmt.Fprintln(os.Stdout, "")
			// }
		}
	}
	<-isFin
}

func read_ignoring_queries(ignoring_queries_filename string) []string {
	ignoring_queries := make([]string, 0, 32)

	if ignoring_queries_filename != "" {
		fp, err := os.Open(ignoring_queries_filename)
		defer fp.Close()

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		reader := bufio.NewReaderSize(fp, 1024)
		for {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			ignoring_queries = append(ignoring_queries, string(line))
		}
	}

	return ignoring_queries
}

func main() {
	nroutine := flag.Int("n", 1, "# of routine")
	ignoring_queries_filename := flag.String("e", "", "text file listing queries to be ignored") // "e" is "excluding" or "exception"
	//ignoring_queries := []string{"and", "of", "from"}
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()

	fp, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n %s [-n INT] [-e FILE] FILE\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer fp.Close()

	var ignoring_queries []string = read_ignoring_queries(*ignoring_queries_filename)

	if *debug {
		fmt.Fprintln(os.Stderr, "debug mode")
		fmt.Fprintln(os.Stderr, "Contents of `ignoring_queries`:")
		for _, s := range ignoring_queries {
			fmt.Fprintln(os.Stderr, s)
		}
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "[%s] accessing zooma api...\n", time.Now().Format("2006-01-02 15:04:05"))

	reader := bufio.NewReaderSize(fp, 8192)
	wg := new(sync.WaitGroup)
	isFin := make(chan interface{}, *nroutine)
	for i := 1; ; i++ {
		if i%1000 == 0 {
			fmt.Fprintf(os.Stderr, "[%s] Sent %d queries\n", time.Now().Format("2006-01-02 15:04:05"), i)
		}
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// split line with \t and obtain query from $3
		// replace special characters with "+"
		split_line := strings.Split(string(line), "\t")
		r := regexp.MustCompile(`[%&!?/.,;*()<>{}\[\]\\^|:#"$=~ ]+`)
		query := r.ReplaceAllString(split_line[2], "+")
		//query := strings.Replace(split_line[2], " ", "+", -1)

		if !is_valid_query(query, ignoring_queries) {
			continue
		}

		wg.Add(1)
		isFin <- struct{}{}
		go get_zooma_json(query, split_line, isFin, wg)
		//decode_data := <- isFin
	}
	wg.Wait()
}
