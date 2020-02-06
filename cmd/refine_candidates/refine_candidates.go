package main

import (
	"fmt"
_	"io/ioutil"
	"io"
	"os"
	"regexp"
	"strings"
	"bufio"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "usage: get_bs_json bsid_list")
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Opening BS ID list...")

	fp, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "Done.")
	defer fp.Close()

	reader := bufio.NewReaderSize(fp, 8192)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sep_line := strings.Split(string(line), "\t")
// - Zooma response の propertyType を cell line と cell type 、またはなしに限る。
//   propertyType がないものに当たりがあるケースもあるので。
// - response の ontology を EFO CLO CL BTO に限る。
// - クエリが単に数値であるもの、"a" "the" "from" "primary" を除く。
//   "293" とかは当たりである可能性もあるが。
// - BioSample 上の attribute 名が "INSDC" を含むものを除く。
		ok := true
		r1 := regexp.MustCompile(`^cell line$|^cell type$|^$`)
		r2 := regexp.MustCompile(`EFO_|CLO_|CL_|BTO_`)
		ok = ok && r1.MatchString(sep_line[3])
		ok = ok && r2.MatchString(sep_line[6])
		if ok {
			//fmt.Fprintln(os.Stdout, sep_line[3])
			fmt.Fprintln(os.Stdout, string(line))
		}
	}
}
