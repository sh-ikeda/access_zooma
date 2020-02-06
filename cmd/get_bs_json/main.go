package main

import (
	"fmt"
	"io/ioutil"
	"io"
	"os"
	"bufio"
	"net/http"
	"flag"
	"strings"
	"sync"
	"time"
)

func get_bs_json(url string, isFirst bool, isFin chan interface{}, wg *sync.WaitGroup, logfile *os.File) bool {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	output := string(bytes)

	if resp.StatusCode >= 400 {
		message := "Error: " + url + "\n" + output + "\n"
		logfile.Write(([]byte)(message))
		<-isFin
		return false
	}

	if !isFirst {
		//fmt.Fprintln(os.Stdout, ",")
		output = ",\n" + output
	}
	fmt.Fprint(os.Stdout, output)

	<-isFin
	return true
}

func main() {
	nroutine := flag.Int("n", 1, "# of routine")

	flag.Parse()

	fp, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "\n", flag.Arg(0))
		fmt.Fprintf(os.Stderr, "Usage of %s:\n %s [-n INT] FILE\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "  FILE\tText file of BioSample ID list\n")
		os.Exit(1)
	}
	defer fp.Close()

	program_name := strings.Replace(os.Args[0], "./", "", -1)
	logfilename := fmt.Sprintf("log.%s.%s.txt", program_name, time.Now().Format("20060102150405"))
	logfile, err := os.Create(logfilename)
    if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
    }
    defer logfile.Close()


	reader := bufio.NewReaderSize(fp, 64)
	wg := new(sync.WaitGroup)
	isFin := make(chan interface{}, *nroutine)
	base := "https://www.ebi.ac.uk/biosamples/samples/"

	fmt.Fprintf(os.Stderr, "[%s] Accessing EBI BioSamples API with %d threads...\n", time.Now().Format("2006-01-02 15:04:05"), *nroutine)

	isFirst := true
	fmt.Fprintln(os.Stdout, "[")
	for i := 1; ; i++ {
		id, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		wg.Add(1)
		isFin <- struct{}{}
		if i == 1 {
			get_bs_json(base + string(id), isFirst, isFin, wg, logfile)
		} else {
			go get_bs_json(base + string(id), isFirst, isFin, wg, logfile)
		}
		if i % 1000 == 0 {
			fmt.Fprintf(os.Stderr, "[%s] Sent %d queries.\n", time.Now().Format("2006-01-02 15:04:05"), i)
		}
	}
	wg.Wait()
	fmt.Fprintln(os.Stdout, "]")

	fmt.Fprintf(os.Stderr, "[%s] Done.\n", time.Now().Format("2006-01-02 15:04:05"))

}
