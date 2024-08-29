package get_jobs

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

type JenkinsResponse struct {
	Class string `json:"_class"`
	Jobs  []Job  `json:"jobs"`
}

type Job struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

func fetchFolder(apiUrl, jenkinsUser, jenkinsPwd string, jobsMap map[string]string) error {
	// 创建请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		logrus.WithError(err).Fatal("error create request")
	}

	// 设置请求头及认证
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Connection", "keep-alive")
	req.SetBasicAuth(jenkinsUser, jenkinsPwd)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Fatal("error make request")
	}
	defer resp.Body.Close()

	resp_data, _ := io.ReadAll(resp.Body)

	//定义一个response的结构体类型来解析json响应
	var response JenkinsResponse
	err = json.Unmarshal(resp_data, &response)
	if err != nil {
		fmt.Println(err)
	}

	for _, job := range response.Jobs {
		processJob(job, jobsMap)
	}
	return nil
}

// 通过class判断job类型,如果是目录则调用fetchFolder遍历目录
func processJob(job Job, jobsMap map[string]string) {
	if job.Class == "com.cloudbees.hudson.plugins.folder.Folder" {
		fetchFolder(job.Url+"api/json?pretty=true", os.Getenv("jenkins_user"), os.Getenv("jenkins_pwd"), jobsMap)
	} else {
		jobsMap[job.Name] = job.Url
	}
}

func GetJobs() map[string]string {
	JenkinsUser := os.Getenv("jenkins_user")
	JenkinsPwd := os.Getenv("jenkins_pwd")
	JenkinsUrl := os.Getenv("jenkins_url")
	apiUrl := JenkinsUrl + "/api/json?pretty=true&tree=jobs[name,url]"

	jobsMap := make(map[string]string)

	err := fetchFolder(apiUrl, JenkinsUser, JenkinsPwd, jobsMap)
	if err != nil {
		logrus.WithError(err).Fatal("error processing jobs")
	}
	fmt.Println(jobsMap)
	return jobsMap
}
