package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

func retrieveMatches(line string, regExp *regexp.Regexp) map[string]string {
	match := regExp.FindStringSubmatch(line)
	matches := make(map[string]string)
	for i, name := range regExp.SubexpNames() {
		if i != 0 && name != "" {
			matches[name] = match[i]
		}
	}

	return matches
}

func main() {
	sourceFilePtr := flag.String("source", "nginx-access.log", "source nginx log file")
	destFilePtr := flag.String("destination", "urls.json", "destination json results file contains grouped top locations")
	//groupBy := flag.String("group-by", "", "field to group by")
	//available for retrieve: count_*, visitors (ip+useragent), ip, date, datetime, method, uri, query, statuscode, bytessent, refferer, useragent
	retrieve := flag.String("get", "", "what fields to retrieve")
	flag.Parse()
	retrieveFields := strings.Split(*retrieve, " ")

	file, _ := ioutil.ReadFile(*sourceFilePtr)
	buf := bytes.NewBuffer(file)
	// todo split datetime on date and time
	sourceLinePattern := `(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<datetime>\d{2}\/[A-Za-z]{3}\/\d{4}:\d{2}:\d{2}:\d{2} (\+|\-)\d{4})\] (("(?P<method>GET|POST|HEAD) )(?P<uri>.+?)(?P<query>\?.*)? (HTTP\/\d\.\d")) (?P<statuscode>\d{3}) (?P<bytessent>\d+) (["](?P<refferer>(\-)|(.*))["]) (["](?P<useragent>.+)["])`
	re := regexp.MustCompile(sourceLinePattern)

	// except insignificant requests
	insignificantOccurrences := `\.woff|\.ttf|\.eot|\.svg|.ico|\.png|\.jpg|\.jpeg|\.gif|\.mp4|\.css\.map|\.js\.map|\.js|\.css|get\-file\?id|robots\.txt|\/admin`
	reInsOcc := regexp.MustCompile(insignificantOccurrences)

	container := make(map[string]map[string]map[string]int)

	for {
		line, err := buf.ReadString('\n')
		if len(line) == 0 {
			if err != nil && err == io.EOF {
				break
			}
		}

		matches := retrieveMatches(line, re)

		// match only GET except insignificantOccurrences
		if matches["method"] == "GET" && !reInsOcc.MatchString(matches["uri"]) {
			date_str := strings.Split(matches["datetime"], ":")[0]
			date, _ := time.Parse("02/Jan/2006", date_str)
			date_str = date.Format("2006-01-02")

			for _, field := range retrieveFields {
				if _, ok := container[date_str]; ok {
					if _, ok2 := container[date_str][field]; ok2 {
						container[date_str][field][matches[field]]++
					} else {
						temp := make(map[string]int)
						temp[matches[field]]++
						container[date_str][field] = temp
					}
				} else {
					temp := make(map[string]map[string]int)
					temp[field] = make(map[string]int)
					temp[field][matches[field]]++
					container[date_str] = temp
				}
			}
		}
	}

	//output, _ := json.Marshal(container)
	output, _ := json.MarshalIndent(container, "", "    ")

	ioutil.WriteFile(*destFilePtr, output, 0644)
}
