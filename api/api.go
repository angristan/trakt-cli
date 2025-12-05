package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type APIClient struct {
	// The url of the API endpoint
	Endpoint string
	// The client for accessing the API
	Client *http.Client
	// The credentials
	Credentials Credentials
}

type Credentials struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	AccessToken  string `yaml:"access-token"`
}

// Create a new API client for the given API version.
func NewAPIClient() APIClient {

	configFile, err := xdg.SearchConfigFile("trakt-cli/config.yaml")
	if err != nil {
		log.Fatalf("Failed to read %q file, please run `trakt auth`", configFile)
	}
	config, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	var creds Credentials
	err = yaml.Unmarshal(config, &creds)
	if err != nil {
		log.Fatalf("Failed to read %q file, please run `trakt auth`", configFile)
	}

	return APIClient{
		Endpoint: "https://api.trakt.tv",
		Client: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout: 5 * time.Second,
			},
		},
		Credentials: creds,
	}
}

type AuthDeviceCodeReq struct {
	ClientID string `json:"client_id"`
}

type AuthDeviceCodeResp struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type requestParams struct {
	method     string
	path       string
	body       interface{}
	auth       bool
	pagination PaginationsParams
	query      map[string]string
}

func (c *APIClient) doRequest(params requestParams) (*http.Response, error) {
	req, err := http.NewRequest(params.method, c.Endpoint+params.path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("trakt-api-version", "2")

	if params.body != nil {
		req.Header.Add("Content-Type", "application/json")
		body, err := json.Marshal(params.body)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	q := req.URL.Query()
	if params.pagination.Page != 0 {
		q.Add("page", fmt.Sprintf("%d", params.pagination.Page))
	}
	if params.pagination.Limit != 0 {
		q.Add("limit", fmt.Sprintf("%d", params.pagination.Limit))
	}
	for k, v := range params.query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	if params.auth {
		req.Header.Add("trakt-api-key", c.Credentials.ClientID)
		req.Header.Add("Authorization", "Bearer "+c.Credentials.AccessToken)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *APIClient) AuthDeviceCode(req *AuthDeviceCodeReq) (*AuthDeviceCodeResp, error) {
	var resp AuthDeviceCodeResp
	httpResp, err := c.doRequest(requestParams{
		method: http.MethodPost,
		path:   "/oauth/device/code",
		body:   req,
		auth:   false,
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

type AuthDeviceTokenReq struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AuthDeviceTokenResp struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int    `json:"created_at"`
}

func (c *APIClient) AuthDeviceToken(req *AuthDeviceTokenReq) (*AuthDeviceTokenResp, error) {
	var resp AuthDeviceTokenResp
	httpResp, err := c.doRequest(requestParams{
		method: http.MethodPost,
		path:   "/oauth/device/token",
		body:   req,
		auth:   false,
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == 200 {
		err = json.NewDecoder(httpResp.Body).Decode(&resp)
		if err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

type UserHistory []HistoryItem

type HistoryItem struct {
	ID        int64     `json:"id"`
	WatchedAt time.Time `json:"watched_at"`
	Action    string    `json:"action"`
	Type      string    `json:"type"`
	Movie     struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Imdb  string `json:"imdb"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"movie,omitempty"`
	Episode struct {
		Season int    `json:"season"`
		Number int    `json:"number"`
		Title  string `json:"title"`
		Ids    struct {
			Trakt  int         `json:"trakt"`
			Tvdb   interface{} `json:"tvdb"`
			Imdb   string      `json:"imdb"`
			Tmdb   int         `json:"tmdb"`
			Tvrage interface{} `json:"tvrage"`
		} `json:"ids"`
	} `json:"episode,omitempty"`
	Show struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt  int         `json:"trakt"`
			Slug   string      `json:"slug"`
			Tvdb   int         `json:"tvdb"`
			Imdb   string      `json:"imdb"`
			Tmdb   int         `json:"tmdb"`
			Tvrage interface{} `json:"tvrage"`
		} `json:"ids"`
	} `json:"show,omitempty"`
}

type PaginationsParams struct {
	Page  int
	Limit int
}

type Pagination struct {
	Page      string `json:"page"`
	Limit     string `json:"limit"`
	PageCount string `json:"page_count"`
	ItemCount string `json:"item_count"`
}

func (c *APIClient) GetUserHistory(user string, params PaginationsParams) (UserHistory, Pagination, error) {
	httpResp, err := c.doRequest(requestParams{
		method:     http.MethodGet,
		path:       fmt.Sprintf("/users/%s/history", user),
		body:       nil,
		auth:       true,
		pagination: params,
	})
	if err != nil {
		return nil, Pagination{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, Pagination{}, fmt.Errorf("failed to get user history: %s", httpResp.Status)
	}

	var resp UserHistory
	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return nil, Pagination{}, err
	}

	pagination := Pagination{
		Page:      httpResp.Header.Get("X-Pagination-Page"),
		Limit:     httpResp.Header.Get("X-Pagination-Limit"),
		PageCount: httpResp.Header.Get("X-Pagination-Page-Count"),
		ItemCount: httpResp.Header.Get("X-Pagination-Item-Count"),
	}

	return resp, pagination, nil
}

type UserSettings struct {
	User struct {
		Username string `json:"username"`
		Private  bool   `json:"private"`
		Name     string `json:"name"`
		Vip      bool   `json:"vip"`
		VipEp    bool   `json:"vip_ep"`
		Ids      struct {
			Slug string `json:"slug"`
			UUID string `json:"uuid"`
		} `json:"ids"`
		JoinedAt time.Time `json:"joined_at"`
		Location string    `json:"location"`
		About    string    `json:"about"`
		Gender   string    `json:"gender"`
		Age      int       `json:"age"`
		Images   struct {
			Avatar struct {
				Full string `json:"full"`
			} `json:"avatar"`
		} `json:"images"`
		VipOg    bool `json:"vip_og"`
		VipYears int  `json:"vip_years"`
	} `json:"user"`
	Account struct {
		Timezone   string `json:"timezone"`
		DateFormat string `json:"date_format"`
		Time24Hr   bool   `json:"time_24hr"`
		CoverImage string `json:"cover_image"`
	} `json:"account"`
	Connections struct {
		Facebook bool `json:"facebook"`
		Twitter  bool `json:"twitter"`
		Google   bool `json:"google"`
		Tumblr   bool `json:"tumblr"`
		Medium   bool `json:"medium"`
		Slack    bool `json:"slack"`
		Apple    bool `json:"apple"`
	} `json:"connections"`
	SharingText struct {
		Watching string `json:"watching"`
		Watched  string `json:"watched"`
		Rated    string `json:"rated"`
	} `json:"sharing_text"`
}

func (c *APIClient) GetUserSettings() (UserSettings, error) {
	httpResp, err := c.doRequest(requestParams{
		method: http.MethodGet,
		path:   "/users/settings",
		body:   nil,
		auth:   true,
	})
	if err != nil {
		return UserSettings{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return UserSettings{}, fmt.Errorf("failed to get user settings: %s", httpResp.Status)
	}

	var resp UserSettings
	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return UserSettings{}, err
	}

	return resp, nil
}

type SearchResult struct {
	Type  string  `json:"type"`
	Score float64 `json:"score"`
	Movie *struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Imdb  string `json:"imdb"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"movie,omitempty"`
	Show *struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Ids   struct {
			Trakt int    `json:"trakt"`
			Slug  string `json:"slug"`
			Tvdb  int    `json:"tvdb"`
			Imdb  string `json:"imdb"`
			Tmdb  int    `json:"tmdb"`
		} `json:"ids"`
	} `json:"show,omitempty"`
}

func (c *APIClient) Search(query string, searchType string) ([]SearchResult, error) {
	httpResp, err := c.doRequest(requestParams{
		method: http.MethodGet,
		path:   "/search/" + searchType,
		auth:   true,
		query: map[string]string{
			"query": query,
		},
		pagination: PaginationsParams{
			Limit: 10,
		},
	})
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("search failed: %s", httpResp.Status)
	}

	var resp []SearchResult
	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
