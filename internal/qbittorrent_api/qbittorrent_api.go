package qbittorrentapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const (
	apiURL = "/api/v2"
	/// Methods
	authMethod     = "/auth"
	appMethod      = "/app"
	torrentsMethod = "/torrents"
	/// Endpoints
	loginEndpoint   = "/login"
	versionEndpoint = "/version"
	infoEndpoint    = "/info"
)

type QbittorrentAPIClient struct {
	client  *http.Client
	baseURL string
}

type TorrentInfoResponse struct {
	Name     string `json:"name"`
	Progress string `json:"progress"`
	Size     string `json:"size"`
	State    string `json:"state"`
	Dlspeed  string `json:"dlspeed"`
}

func NewAPIClient(baseURL string) *QbittorrentAPIClient {
	jar, _ := cookiejar.New(nil)
	return &QbittorrentAPIClient{
		client: &http.Client{
			Jar: jar,
		},
		baseURL: baseURL,
	}

}

func (c *QbittorrentAPIClient) Login(username, password string) error {
	var err error

	reqURL := c.baseURL + apiURL + authMethod + loginEndpoint
	params := url.Values{}

	params.Add("username", username)
	params.Add("password", password)

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	cookie := resp.Cookies()
	if len(cookie) < 1 {
		return fmt.Errorf("Error getting cookie")
	}

	return err
}

func (c *QbittorrentAPIClient) GetVersion() (string, error) {
	var err error
	reqURL := c.baseURL + apiURL + appMethod + versionEndpoint

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	version := strings.TrimSpace(string(bodyBytes))

	return version, err
}

func (c *QbittorrentAPIClient) GetTorrentList() ([]TorrentInfoResponse, error) {
	var err error
	torrentList := []TorrentInfoResponse{}

	reqURL := c.baseURL + apiURL + torrentsMethod + infoEndpoint

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return torrentList, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return torrentList, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&torrentList)
	if err != nil {
		return torrentList, err
	}

	return torrentList, nil
}
