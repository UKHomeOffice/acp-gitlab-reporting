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
	gitlabHost           = flag.String("gitlab-host", "", "Gitlab host API")
	gitlabAccessToken    = flag.String("gitlab-access-token", "", "Gitlab access token used to authenticate against the API.")
	reportingUrl         = flag.String("reporting-url", "", "Reporting endpoint url.")
	reportingAccessToken = flag.String("reporting-access-token", "", "Access token used to authenticate against the reporting API.")
	dryRun               = flag.Bool("dry-run", false, "Flag if true will not send the report to the remote endpoint.")
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

	// Loops until a response doesn't return an X-Next-Page header
	for {
		req, err := http.NewRequest("GET", *gitlabHost+"/api/v4/projects", nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		req.Header.Add("PRIVATE-TOKEN", *gitlabAccessToken)
		req.URL.RawQuery = "per_page=100&pagination=true&page=" + strconv.Itoa(page)
		response_body, header := doHttpRequest(req)

		var projects []GitlabProject

		if err := json.Unmarshal(response_body, &projects); err != nil {
			log.Fatal(err.Error())
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
		log.Fatal(err.Error())
	}

	log.Println("Report:")
	log.Println(string(report_json))

	// If this isn't a dry run, publish the report
	if !*dryRun {
		reporting_req, err := http.NewRequest("PUT", *reportingUrl, bytes.NewBuffer(report_json))
		if err != nil {
			log.Fatal(err.Error())
		}
		reporting_req.Header.Add("x-api-key", *reportingAccessToken)
		body, _ := doHttpRequest(reporting_req)
		log.Println("Response payload: ")
		log.Println(body)
	}
}

// doHttpRequest do http request and return body and http headers
func doHttpRequest(req *http.Request) ([]byte, http.Header) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err.Error())
	}

	if resp.StatusCode != 200 {
		log.Println(string(body))
		log.Fatal("Unexpected status code")
	}
	return body, resp.Header
}
