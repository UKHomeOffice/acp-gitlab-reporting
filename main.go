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

func getGitlabStatistics() []byte {
	req, _ := http.NewRequest("GET", *gitlab_url+"/api/v4/application/statistics", nil)
	req.Header.Add("PRIVATE-TOKEN", *gitlab_access_token)

	response_body, _ := http_request(req)
	return response_body
}

func extractGitlabStatistics(response []byte) (int64, int64) {
	var application_statistics map[string]string
	if err := json.Unmarshal(response, &application_statistics); err != nil {
		panic(err)
	}

	total, err := strconv.ParseInt(strings.Replace(application_statistics["projects"], ",", "", -1), 10, 64)
	if err != nil {
		panic(err)
	}

	forks, err := strconv.ParseInt(strings.Replace(application_statistics["forks"], ",", "", -1), 10, 64)
	if err != nil {
		panic(err)
	}
	return total, forks
}

func main() {
	flag.Parse()

	gitlab_statistics_repsonse := getGitlabStatistics()
	total, forks := extractGitlabStatistics(gitlab_statistics_repsonse)

	report_json, err := json.Marshal(report_payload{
		Total:    total,
		Forks:    forks,
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
	body, _ := http_request(reporting_req)
	fmt.Println(body)

}

type application_statistics struct {
}

type report_payload struct {
	Total    int64 `json:"total"`    // The total number of repos
	Forks    int64 `json:"forks"`    // The number of forks
	Archived int64 `json:"archived"` // The number of archived/unused repos if possible
	Personal int64 `json:"personal"` // The number of personal user repos
	Groups   int64 `json:"groups"`   // The number of GitLab groups repos
}

func http_request(req *http.Request) ([]byte, http.Header) {
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
	return body, resp.Header
}
