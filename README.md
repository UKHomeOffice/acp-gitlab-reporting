 # acp-gitlab-reporting
 
 This is an internal reporting service to send Gitlab repository statistics to the Tooling team.
 
 ## Usage
 
 ``` 
./acp-gitlab-reporter 
 ```
 
| Parameter      | Description | Default      | Required |
| ----------- | ----------- | ----------- | ----------- |
| -dry-run      | Flag if true will not send the report to the remote endpoint.       | false      | false       |
|  -gitlab-access-token      | Gitlab access token used to authenticate against the API.       | n/a      | true       |
|  -gitlab-host      | Gitlab host API      | n/a      | true       |
| -reporting-access-token      | Access token used to authenticate against the reporting API.       | n/a      | true       |
| -reporting-url      | Reporting URL.      | n/a      | true       |
 
 
 ## Report payload structure
 
 
 ```
 {
  total: number;       // The total number of repos
  forks: number;       // The number of forks
  archived: number;    // The number of archived/unused repos if possible
  personal: number;    // The number of personal user repos
  groups: number;      // The number of GitLab groups repos
}
 ```
