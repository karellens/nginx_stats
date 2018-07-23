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
    "crypto/md5"
    "encoding/hex"
)


func main() {
    sourceFilePtr := flag.String("source", "nginx-access.log", "source nginx log file")
    destFilePtr := flag.String("destination", "stats.json", "destination json results file")
    flag.Parse()

    file, _ := ioutil.ReadFile(*sourceFilePtr)

    buf := bytes.NewBuffer(file)

    container := make( map[string]map[string]int )

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

        // except insignificant requests
        var re_is = regexp.MustCompile(`\.woff|\.ttf|\.eot|\.svg|.ico|\.png|\.jpg|\.jpeg|\.gif|\.mp4|\.css\.map|\.js\.map|\.js|\.css|get\-file\?id|robots\.txt|\/admin`)
        if matches["method"]=="GET" && !re_is.MatchString(matches["url"]) {
            date_str := strings.Split(matches["dateandtime"], ":")[0]
            date, _ := time.Parse("02/Jan/2006", date_str)
            date_str = date.Format("2006-01-02")

            h := md5.New()
            io.WriteString(h, matches["ipaddress"]+matches["useragent"])
            hash := hex.EncodeToString(h.Sum(nil))

            if _, ok := container[date_str]; ok {
                container[date_str][hash]++
            } else {
                temp := make(map[string]int)
                temp[hash]++
                container[date_str] = temp
            }
        }
    }

    output, _ := json.Marshal(container)

    ioutil.WriteFile(*destFilePtr, output, 0644)
}
