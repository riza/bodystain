package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uilive"
	"github.com/parnurzeal/gorequest"
)

var (
	containKey, inputfile, output, port string
	done                                chan bool
	buf                                 bytes.Buffer

	scanner *bufio.Scanner
	timeout = 1000 * time.Millisecond
	mutex   sync.Mutex
)

const (
	Green  = "%s\033[1;32m%s\033[0m %s\n"
	Cyan   = "%s\033[1;36m%s\033[0m %s\n"
	Yellow = "%s\033[1;33m%s\033[0m %s\n"
	Red    = "%s\033[1;31m%s\033[0m %s\n"
)

func init() {
	flag.StringVar(&inputfile, "input", "", "-input=file")
	flag.StringVar(&output, "output", "", "-output=file")
	flag.StringVar(&port, "port", "", "-port=3001")
	flag.StringVar(&containKey, "contains", "", "-contains=Claymore")
}

func main() {

	flag.Parse()

	inputFile, err := os.Open(inputfile)

	if err != nil {
		log.Fatal(err)
	}

	outputFile, err := initOutputFile(output)

	if err != nil {
		errmsg(err, true)
	}

	defer outputFile.Close()

	tee := io.TeeReader(inputFile, &buf)
	count, err := lineCounter(tee)

	if err != nil {
		errmsg(err, true)
	}

	writer := uilive.New()
	writer.Start()

	info("BODYstain v1.0b - Automated Key Searcher", false)

	scanner = bufio.NewScanner(&buf)
	done = make(chan bool)

	infoStr := fmt.Sprintf("Total URL: %d", count)

	info(infoStr, false)
	current := 0
	found := 0
	go func(done chan bool) {
		for scanner.Scan() {

			url := fmt.Sprintf("http://%s:%s", scanner.Text(), port)
			_, body, errs := gorequest.New().Timeout(timeout).Get(url).End()

			if errs != nil {
				/* 	for _, err := range errs {
					info(err.Error(), false)
				} */

			}

			current++

			contains := strings.Contains(body, containKey)

			if contains {
				found++
				txt := fmt.Sprintf("Your key contains response: %s", url)
				green(txt, true)

				err = writeLine(outputFile, url)

				if err != nil {
					errmsg(err, true)
				}

			}

			//fmt.Fprintf(writer, "Checking.. (%d/%d) URL - Found %d\n", current, count, found)

		}

		fmt.Fprintln(writer, "Check sucessfully: Found %d/%d", found, count)
		writer.Stop()

		done <- true
	}(done)

	<-done
}

//utils

func info(msg string, nl bool) {
	var prefix string

	if nl {
		prefix = "\n"
	}

	fmt.Printf(Yellow, prefix, "[MSG] ", msg)
}

func green(host string, nl bool) {
	var prefix string

	if nl {
		prefix = "\n"
	}

	fmt.Printf(Green, prefix, "[LIVE] ", host)
}

func errmsg(errmsg error, nl bool) {
	var prefix string

	if nl {
		prefix = "\n"
	}

	fmt.Printf(Red, prefix, "[ERR] ", errmsg)
	os.Exit(1)
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 1
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {

			break
		}

		if err != nil {
			return count, err
		}

	}

	return count, nil

}

func fileExists(filename string) (file *os.File, exists bool, err error) {

	if _, err = os.Stat(output); os.IsNotExist(err) {
		return nil, false, nil
	} else {
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		return file, true, err
	}
}

func createFile(filename string) (outputFile *os.File, err error) {

	file, err := os.Create(filename)

	if err != nil {
		return
	}

	defer file.Close()

	outputFile, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	if err != nil {
		return
	}

	return
}

func initOutputFile(output string) (outputFile *os.File, err error) {

	outputFile, exists, err := fileExists(output)

	if err != nil {
		return
	}

	if !exists {
		outputFile, err = createFile(output)

		if err != nil {
			return
		}
	} else {
		outputFile, err = os.OpenFile(output, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		if err != nil {
			return
		}
	}

	return
}

func writeLine(file *os.File, host string) error {

	mutex.Lock()
	defer mutex.Unlock()

	_, err := file.WriteString(host + "\n")

	if err != nil {
		return err
	}

	return nil
}
