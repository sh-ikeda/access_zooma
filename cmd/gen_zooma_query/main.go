package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func contains(a []string, s string) bool {
	for _, q := range a {
		if s == q {
			return true
		}
	}
	return false
}

func parseJSON(bytes []byte, d *interface{}) {
	fmt.Fprintln(os.Stderr, "Parsing...")
	if err := json.Unmarshal(bytes, d); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if err, ok := err.(*json.SyntaxError); ok {
			fmt.Fprintln(os.Stderr, string(bytes[err.Offset-15:err.Offset+15]))
		}
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Done. (Total %d entries)\n", len((*d).([]interface{})))
}

func main() {
	a := flag.String("a", "", "comma-separated string of attributes to be collected")
	s := flag.Bool("s", false, "If true, a value of attributes not listed in -a option is separated with whitespace and each word is output as a query")
	l := flag.Bool("l", false, "If true, all attributes are output for entries without any attributes listed in -a")
	flag.Parse()

	attr := strings.Split(*a, ",")
	outputAll := *l
	toBeSeparated := *s

	bytes, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n %s [-a LIST] [-s] FILE\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var decode_data interface{}
	parseJSON(bytes, &decode_data)

	// Collect designated attributes and output them
	fmt.Fprintln(os.Stderr, "Writing...")
	for _, data := range decode_data.([]interface{}) {
		d := data.(map[string]interface{})
		id := d["accession"].(string)
		tax := d["taxId"].(float64)
		if int(tax) != 9606 {
			continue
		}
		ch := d["characteristics"].(map[string]interface{})
		n := 0
		for key, _ := range ch {
			if contains(attr, key) || *a == "" {
				c := ch[key].([]interface{})
				x := c[0].(map[string]interface{})
				value := x["text"].(string)
				// in case that value has "\t"
				value = strings.Replace(value, "\t", " ", -1)
				fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t\n", id, key, value)
				n++
			}
		}
		// When no attribute contains any keywords, output all
		if n == 0 && outputAll {
			for key, _ := range ch {
				c := ch[key].([]interface{})
				x := c[0].(map[string]interface{})
				value := x["text"].(string)
				//slice := strings.Split(value, " ")
				slice := regexp.MustCompile("[ (),./]+").Split(value, -1)
				fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t\n", id, key, value)
				if toBeSeparated && len(slice) > 1 {
					for _, word := range slice {
						if word != "" {
							fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%s\n", id, key, word, value)
						}
					}
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "Done.")
}
