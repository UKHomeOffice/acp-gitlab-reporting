package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	gitlab_url             = flag.String("gitlab_url", "", "")
	gitlab_access_token    = flag.String("gitlab_access_token", "", "")
	reporting_url          = flag.String("reporting_url", "", "")
	reporting_access_token = flag.String("reporting_access_token", "", "")
)

func main() {
	flag.Parse()

	req, _ := http.NewRequest("GET", *gitlab_url+"/api/v4/application/statistics", nil)
	req.Header.Add("PRIVATE-TOKEN", *gitlab_access_token)

	var application_statistics map[string]string
	if err := json.Unmarshal(http_request(req), &application_statistics); err != nil {
		panic(err)
	}

	total, err := strconv.ParseInt(strings.Replace(application_statistics["projects"], ",", "", -1), 10, 32)
	if err != nil {
		panic(err)
	}

	forks, err := strconv.ParseInt(strings.Replace(application_statistics["forks"], ",", "", -1), 10, 32)
	if err != nil {
		panic(err)
	}

	report_json, err := json.Marshal(report_payload{
		Total:    int(total),
		Forks:    int(forks),
		Archived: 0,
		Personal: 0,
		Groups:   0,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(string(report_json))

	reporting_req, _ := http.NewRequest("PUT", *reporting_url, bytes.NewBuffer(report_json))
	reporting_req.Header.Add("x-api-key", *reporting_access_token)
	body := http_request(reporting_req)
	fmt.Println(body)

}

type application_statistics struct {
}

type report_payload struct {
	Total    int `json:"total"`    // The total number of repos
	Forks    int `json:"forks"`    // The number of forks
	Archived int `json:"archived"` // The number of archived/unused repos if possible
	Personal int `json:"personal"` // The number of personal user repos
	Groups   int `json:"groups"`   // The number of GitLab groups repos
}

func http_request(req *http.Request) []byte {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic("Incorrect status code")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	return body
}
