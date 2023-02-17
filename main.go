package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	gitlab_url             = flag.String("gitlab_url", "", "")
	gitlab_access_token    = flag.String("gitlab_access_token", "", "")
	reporting_url          = flag.String("reporting_url", "", "")
	reporting_access_token = flag.String("reporting_access_token", "", "")
)

type Project struct {
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
		fmt.Println(string(body))
		panic("Incorrect status code")
	}
	return body, resp.Header
}

func main() {
	flag.Parse()

	var total, forks, archived, personal, groups = 0, 0, 0, 0, 0
	var page = 1

	for {
		req, err := http.NewRequest("GET", *gitlab_url+"/api/v4/projects", nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("PRIVATE-TOKEN", *gitlab_access_token)
		req.URL.RawQuery = "per_page=100&pagination=true&page=" + strconv.Itoa(page)
		response_body, header := http_request(req)

		var projects []Project

		if err := json.Unmarshal(response_body, &projects); err != nil {
			panic(err)
		}

		total += len(projects)

		for _, project := range projects {
			forks += project.Forks_count
			if project.Archived == true {
				archived++
			}
			if project.Namespace.Kind == "user" {
				personal++
			} else if project.Namespace.Kind == "group" {
				groups++
			}
		}

		if header.Get("X-Next-Page") != "" {
			page, _ = strconv.Atoi(header.Get("X-Next-Page"))
			continue
		}

		break
	}

	report_json, err := json.Marshal(ReportPayload{
		Total:    total,
		Forks:    forks,
		Archived: archived,
		Personal: personal,
		Groups:   groups,
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
