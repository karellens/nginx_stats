package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

func retrieveMatches(line string, regExp *regexp.Regexp) (map[string]string, error) {
	var err error
	err = nil
	defer func() {
		e := recover()
		if panicErr, ok := e.(error); ok {
			panicErr = panicErr
			return
		}
	}()

	match := regExp.FindStringSubmatch(line)
	matches := make(map[string]string)
	for i, name := range regExp.SubexpNames() {
		if i != 0 && name != "" {
			matches[name] = match[i]
		}
	}

	return matches, err
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func main() {
	// Available for retrieve: ip, date, datetime, method, uri, query, statuscode, bytessent, refferer, useragent.
	// You can combine the above fields with `+` to calculate the number of occurrences of unique combinations
	// Specify `=` before the combination if you want to display only the number of unique occurrences found
	sourceFilePtr := flag.String("source", "nginx-access.log", "source nginx log file")
	destFilePtr := flag.String("destination", "urls.json", "destination json results file contains grouped top locations")
	fromDatePtr := flag.String("from", "", "from date")
	toDatePtr := flag.String("to", "", "to date")
	prettyPtr := flag.Bool("pretty", false, "Pretty JSON output")
	retrievePtr := flag.String("get", "", "what fields to retrieve")
	flag.Parse()
	retrieveFields := strings.Split(*retrievePtr, " ")

	file, _ := ioutil.ReadFile(*sourceFilePtr)
	buf := bytes.NewBuffer(file)

	fromDate, _ := time.Parse("2006-01-02", *fromDatePtr)
	var toDate time.Time
	if len(*toDatePtr) == 0 {
		toDate = time.Now()
	} else {
		toDate, _ = time.Parse("2006-01-02", *toDatePtr)
	}

	// TODO: split datetime on date and time
	sourceLinePattern := `(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<datetime>\d{2}\/[A-Za-z]{3}\/\d{4}:\d{2}:\d{2}:\d{2} (\+|\-)\d{4})\] (("(?P<method>GET|POST|HEAD) )(?P<uri>.+?)(?P<query>\?.*)? (HTTP\/\d\.\d")) (?P<statuscode>\d{3}) (?P<bytessent>\d+) (["](?P<refferer>(\-)|(.*))["]) (["](?P<useragent>.+)["])`
	re := regexp.MustCompile(sourceLinePattern)

	// except insignificant requests. You can add your own patterns
	insignificantOccurrences := `\.woff|\.ttf|\.eot|\.svg|.ico|\.png|\.jpg|\.jpeg|\.gif|\.mp4|\.css\.map|\.js\.map|\.js|\.css|get\-file\?id|robots\.txt|\/admin`
	reInsOcc := regexp.MustCompile(insignificantOccurrences)

	container := make(map[string]map[string]map[string]int)

	var clean_field string
	var field_value string
	var combined_fields []string

	for {
		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err != nil && err == io.EOF {
				break
			}
		}

		matches, err := retrieveMatches(line, re)
		if err != nil {
			continue
		}

		date_str := strings.Split(matches["datetime"], ":")[0]
		date, _ := time.Parse("02/Jan/2006", date_str)
		if !inTimeSpan(fromDate, toDate, date) {
			continue
		}

		// match only GET except insignificantOccurrences
		if matches["method"] == "GET" && !reInsOcc.MatchString(matches["uri"]) {
			date_str = date.Format("2006-01-02")

			for _, field := range retrieveFields {

				if strings.Contains(field, "+") {
					if strings.HasPrefix(field, "=") {
						clean_field = field[1:]
					} else {
						clean_field = field
					}
					combined_fields = strings.Split(clean_field, "+")
					var unpacked_field_value string

					for _, cf := range combined_fields {
						unpacked_field_value += matches[cf]
					}
					hash := sha1.Sum([]byte(unpacked_field_value))
					field_value = base32.HexEncoding.EncodeToString(hash[:])

				} else {
					field_value = matches[field]
				}

				if _, ok := container[date_str]; ok {
					if _, ok2 := container[date_str][field]; ok2 {
						container[date_str][field][field_value]++
					} else {
						temp := make(map[string]int)
						temp[field_value]++
						container[date_str][field] = temp
					}
				} else {
					temp := make(map[string]map[string]int)
					temp[field] = make(map[string]int)
					temp[field][field_value]++
					container[date_str] = temp
				}
			}
		}
	}

	for _, field := range retrieveFields {
		if strings.HasPrefix(field, "=") {
			for _, day_block := range container {
				temp := make(map[string]int)
				temp["total"] = len(day_block[field])
				day_block[field] = temp
			}
		}
	}

	var output []byte
	if *prettyPtr {
		output, _ = json.MarshalIndent(container, "", "    ")
	} else {
		output, _ = json.Marshal(container)
	}

	ioutil.WriteFile(*destFilePtr, output, 0644)
}
