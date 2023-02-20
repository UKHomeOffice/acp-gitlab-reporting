package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var (
	gitlabUrl            = flag.String("gitlab-url", "", "")
	gitlabAccessToken    = flag.String("gitlab-access-token", "", "")
	reportingUrl         = flag.String("reporting-url", "", "")
	reportingAccessToken = flag.String("reporting-access-token", "", "")
	dryRun               = flag.Bool("dry-run", false, "")
)

type GitlabProject struct {
	Forks_count int  `json:"forks_count"`
	Archived    bool `json:"archived"`
	Namespace   struct {
		Kind string `json:"kind"`
	} `json:"namespace"`
}

type ReportPayload struct {
	Total    int `json:"total"`    // The total number of repos
	Forks    int `json:"forks"`    // The number of forks
	Archived int `json:"archived"` // The number of archived/unused repos if possible
	Personal int `json:"personal"` // The number of personal user repos
	Groups   int `json:"groups"`   // The number of GitLab groups repos
}

func main() {
	flag.Parse()

	report := ReportPayload{
		Total:    0,
		Forks:    0,
		Archived: 0,
		Personal: 0,
		Groups:   0,
	}
	var page = 1

	for {
		req, err := http.NewRequest("GET", *gitlabUrl+"/api/v4/projects", nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("PRIVATE-TOKEN", *gitlabAccessToken)
		req.URL.RawQuery = "per_page=100&pagination=true&page=" + strconv.Itoa(page)
		response_body, header := http_request(req)

		var projects []GitlabProject

		if err := json.Unmarshal(response_body, &projects); err != nil {
			panic(err)
		}

		report.Total += len(projects)

		for _, project := range projects {
			report.Forks += project.Forks_count
			if project.Archived {
				report.Archived++
			}
			if project.Namespace.Kind == "user" {
				report.Personal++
			} else if project.Namespace.Kind == "group" {
				report.Groups++
			}
		}

		if header.Get("X-Next-Page") != "" {
			page, _ = strconv.Atoi(header.Get("X-Next-Page"))
			continue
		}

		break
	}

	report_json, err := json.Marshal(report)

	if err != nil {
		panic(err)
	}

	log.Println("Report:")
	log.Println(string(report_json))

	if !*dryRun {
		reporting_req, _ := http.NewRequest("PUT", *reportingUrl, bytes.NewBuffer(report_json))
		reporting_req.Header.Add("x-api-key", *reportingAccessToken)
		body, _ := http_request(reporting_req)
		log.Println("Response payload: ")
		log.Println(body)
	}
}

func http_request(req *http.Request) ([]byte, http.Header) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		log.Println(string(body))
		panic("Incorrect status code")
	}
	return body, resp.Header
}
