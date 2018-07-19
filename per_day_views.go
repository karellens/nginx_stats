package main

import (
    "bytes"
    "io"
    "io/ioutil"
    "regexp"
    "strings"
    "encoding/json"
)


func main() {
    filename := "geely2/storage/logs/nginx-access.log"

    file, _ := ioutil.ReadFile(filename)

    buf := bytes.NewBuffer(file)

    counter := make( map[string]int )

    for {
        line, err := buf.ReadString('\n')
        if len(line) == 0 {
            if err != nil && err == io.EOF {
                break
            }
        }

        re := regexp.MustCompile(`(?P<ipaddress>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}) - - \[(?P<dateandtime>\d{2}\/[A-Za-z]{3}\/\d{4}:\d{2}:\d{2}:\d{2} (\+|\-)\d{4})\] (("(GET|POST|HEAD) )(?P<url>.+) (HTTP\/\d\.\d")) (?P<statuscode>\d{3}) (?P<bytessent>\d+) (["](?P<refferer>(\-)|(.*))["]) (["](?P<useragent>.+)["])`)
        match := re.FindStringSubmatch(line)
        matches := make(map[string]string)
        for i, name := range re.SubexpNames() {
            if i != 0 && name != "" {
                matches[name] = match[i]
            }
        }

        date := strings.Split(matches["dateandtime"], ":")[0]

        // except insignificant requests
        var re_is = regexp.MustCompile(`\.woff|\.ttf|\.eot|\.svg|.ico|\.png|\.jpg|\.jpeg|\.gif|\.mp4|\.css\.map|\.js\.map|\.js|\.css|get\-file\?id|robots\.txt`)
        if !re_is.MatchString(matches["url"]) {
            counter[date+"-"+matches["ipaddress"]]++
        }
    }

    results := make(map[string]map[string]int)
    for k, v := range counter {
        date_ip := strings.Split(k, "-")
        date, ip := date_ip[0], date_ip[1]

        if _, ok := results[date]; ok {
            results[date][ip] = v
        } else {
            temp := make(map[string]int)
            temp[ip] = v
            results[date] = temp
        }
    }

    output, _ := json.Marshal(results)

    ioutil.WriteFile("stats.json", output, 0644)
}
