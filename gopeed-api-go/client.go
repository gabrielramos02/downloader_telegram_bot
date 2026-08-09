package gopeed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	apiVersion      = "v1"
	infoEndpoint    = "/info"
	taskEndpoint    = "/tasks"
	resolveEndpoint = "/resolve"
)

type GopeedClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (*GopeedClient, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}
	url.Path = "api/" + apiVersion
	httpClient := &http.Client{}
	return &GopeedClient{
		baseURL:    url.String(),
		httpClient: httpClient,
	}, nil
}

func (c *GopeedClient) GetInfo(path string) (GopeedInfo, error) {
	var resp GopeedResponse[GopeedInfo]
	req, _ := http.NewRequest("GET", c.baseURL+infoEndpoint, nil)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return resp.Data, fmt.Errorf("failed to get info: %v", err)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return resp.Data, fmt.Errorf("failed to decode response from GetInf from GetInfoo: %v", err)
	}
	if resp.Code != 0 {
		return resp.Data, fmt.Errorf("error from server: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *GopeedClient) GetTasks() ([]GopeedTask, error) {
	var resp GopeedResponse[[]GopeedTask]
	req, _ := http.NewRequest("GET", c.baseURL+taskEndpoint, nil)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return resp.Data, fmt.Errorf("failed to get taks: %v", err)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return resp.Data, fmt.Errorf("failed to decode response from getTasks: %v", err)
	}
	if resp.Code != 0 {
		return resp.Data, fmt.Errorf("error from server: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *GopeedClient) GetTask(taskID string) (GopeedTask, error) {
	var resp GopeedResponse[GopeedTask]
	req, _ := http.NewRequest("GET", c.baseURL+taskEndpoint+"/"+taskID, nil)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return resp.Data, fmt.Errorf("failed to get taks: %v", err)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return resp.Data, fmt.Errorf("failed to decode response from GetTask: %v", err)
	}
	if resp.Code != 0 {
		return resp.Data, fmt.Errorf("error from server: %s", resp.Msg)
	}
	return resp.Data, nil
}
func (c *GopeedClient) CreateTask(url string) (taskid string, err error) {
	var resp GopeedResponse[string]
	resolvedID, err := c.resolve(url)
	if err != nil {
		return "", err
	}
	payload := GopeedCreateTask{
		Rid: resolvedID,
		Opts: GopeedOptions{
			Extra: &GopeedExtraOptions{Connections: 32},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}
	reader := bytes.NewReader(jsonData)
	req, _ := http.NewRequest("POST", c.baseURL+taskEndpoint, reader)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %v", err)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response from createTask: %v", err)
	}
	if resp.Code != 0 {
		return resp.Data, fmt.Errorf("error from server: %s", resp.Msg)
	}
	return resp.Data, nil

}
func (c *GopeedClient) resolve(url string) (resolvedID string, err error) {
	var resp GopeedResponse[GopeedResolved]
	payload := GopeedResolve{
		Req: GopeedRequest{URL: url},
		Opts: GopeedOptions{
			Extra: &GopeedExtraOptions{Connections: 32},
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}
	reader := bytes.NewReader(jsonData)

	req, _ := http.NewRequest("POST", c.baseURL+resolveEndpoint, reader)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp.Data.Id, fmt.Errorf("failed to resolve resource: %v", err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return resp.Data.Id, fmt.Errorf("failed to decode response from resolve: %v", err)
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("error from server: %s", resp.Msg)
	}
	return resp.Data.Id, nil
}

func (c *GopeedClient) DeleteTask(taskID string) error {
	var resp GopeedResponse[struct{}]
	params := url.Values{}
	params.Add("force", "true")
	req, _ := http.NewRequest("DELETE", c.baseURL+taskEndpoint+"/"+taskID+"?"+params.Encode(), nil)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("error from server: %s", resp.Msg)
	}
	return nil
}
