package main

import (
    "bytes"
    "io"
    "io/ioutil"
    "regexp"
    "strings"
    "encoding/json"
    "time"
    "flag"
)


func main() {
    sourceFilePtr := flag.String("source", "nginx-access.log", "source nginx log file")
    destFilePtr := flag.String("destination", "stats.json", "destination json results file")
    flag.Parse()

    file, _ := ioutil.ReadFile(*sourceFilePtr)

    buf := bytes.NewBuffer(file)

    counter := make( map[string]int )

    for {
        line, err := buf.ReadString('\n')
        if len(line) == 0 {
            if err != nil && err == io.EOF {
                break
            }
        }

        re := regexp.MustCompile(`(?P<ipaddress>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<dateandtime>\d{2}\/[A-Za-z]{3}\/\d{4}:\d{2}:\d{2}:\d{2} (\+|\-)\d{4})\] (("(?P<method>GET|POST|HEAD) )(?P<url>.+) (HTTP\/\d\.\d")) (?P<statuscode>\d{3}) (?P<bytessent>\d+) (["](?P<refferer>(\-)|(.*))["]) (["](?P<useragent>.+)["])`)
        match := re.FindStringSubmatch(line)
        matches := make(map[string]string)
        for i, name := range re.SubexpNames() {
            if i != 0 && name != "" {
                matches[name] = match[i]
            }
        }

        date := strings.Split(matches["dateandtime"], ":")[0]

        // except insignificant requests
        var re_is = regexp.MustCompile(`\.woff|\.ttf|\.eot|\.svg|.ico|\.png|\.jpg|\.jpeg|\.gif|\.mp4|\.css\.map|\.js\.map|\.js|\.css|get\-file\?id|robots\.txt|\/admin`)
        if matches["method"]=="GET" && !re_is.MatchString(matches["url"]) {
            counter[date+"-"+matches["ipaddress"]]++
        }
    }

    results := make(map[string]map[string]int)
    for k, v := range counter {
        date_ip := strings.Split(k, "-")
        date_str, ip := date_ip[0], date_ip[1]
        date, _ := time.Parse("02/Jan/2006", date_str)
        date_str = date.Format("2006-01-02")

        if _, ok := results[date_str]; ok {
            results[date_str][ip] = v
        } else {
            temp := make(map[string]int)
            temp[ip] = v
            results[date_str] = temp
        }
    }

    output, _ := json.Marshal(results)

    ioutil.WriteFile(*destFilePtr, output, 0644)
}
