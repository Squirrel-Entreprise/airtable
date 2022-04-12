package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client  HTTPClient
	attempt int
)

func init() {
	Client = &http.Client{
		Timeout: time.Second * 10,
	}
}

const (
	apiUrl = "https://api.airtable.com/v0"
)

type Airtable struct {
	apiKey string
	base   string
}

func New(apiKey, base string) *Airtable {
	return &Airtable{
		apiKey: apiKey,
		base:   base,
	}
}

type Table struct {
	Name       string `json:"name"`       // table name
	MaxRecords string `json:"maxRecords"` // max 100
	View       string `json:"view"`       // Grid view
}

func (a *Airtable) List(table Table, response interface{}) error {
	return a.call(GET, table, nil, nil, response)
}

func (a *Airtable) Get(table Table, id string, response interface{}) error {
	return a.call(GET, table, &id, nil, response)
}
func (a *Airtable) Create(table Table, data []byte, response interface{}) error {
	return a.call(POST, table, nil, data, response)
}

func (a *Airtable) Update(table Table, id string, data []byte, response interface{}) error {
	return a.call(PATCH, table, &id, data, response)
}

func (a *Airtable) Delete(table Table, id string) error {
	return a.call(DELETE, table, &id, nil, nil)
}

type methodHttp string

const (
	GET    methodHttp = http.MethodGet
	POST   methodHttp = http.MethodPost
	PUT    methodHttp = http.MethodPut
	PATCH  methodHttp = http.MethodPatch
	DELETE methodHttp = http.MethodDelete
)

func (a *Airtable) call(method methodHttp, table Table, id *string, payload []byte, response interface{}) error {

	if table.MaxRecords == "" {
		table.MaxRecords = "100"
	}

	if table.View == "" {
		table.View = "Grid view"
	}

	if table.Name == "" {
		return fmt.Errorf("table name is required")
	}

	table.View = url.QueryEscape(table.View)
	table.Name = url.QueryEscape(table.Name)

	var path string

	// list
	if method == GET && id == nil {
		path = fmt.Sprintf("%s/%s/%s?maxRecords=%s&view=%s", apiUrl, a.base, table.Name, table.MaxRecords, table.View)
	}

	// get || delete || update
	if (method == GET && id != nil) || (method == DELETE && id != nil || (method == PUT && id != nil || method == PATCH && id != nil)) {
		path = fmt.Sprintf("%s/%s/%s/%s", apiUrl, a.base, table.Name, *id)
	}

	// create
	if method == POST {
		path = fmt.Sprintf("%s/%s/%s", apiUrl, a.base, table.Name)
	}

	req, err := http.NewRequest(string(method), path, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	req.Header.Add("Content-Type", "application/json")

	res, err := Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		if attempt < 5 {
			attempt++
			time.Sleep(time.Second * 1)
			return a.call(method, table, id, payload, response)
		}
		return fmt.Errorf("the API is limited to 5 requests per second per base. If you exceed this rate, you will receive a 429 status code and will need to wait 30 seconds before subsequent requests will succeed")
	}

	if res.StatusCode == http.StatusBadRequest {
		return fmt.Errorf("the request encoding is invalid; the request can't be parsed as a valid JSON")
	}

	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("accessing a protected resource without authorization or with invalid credentials")
	}

	if res.StatusCode == http.StatusPaymentRequired {
		return fmt.Errorf("the account associated with the API key making requests hits a quota that can be increased by upgrading the Airtable account plan")
	}

	if res.StatusCode == http.StatusForbidden {
		return fmt.Errorf("accessing a protected resource with API credentials that don't have access to that resource")
	}

	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("route or resource is not found. This error is returned when the request hits an undefined route, or if the resource doesn't exist (e.g. has been deleted)")
	}

	if res.StatusCode == http.StatusRequestEntityTooLarge {
		return fmt.Errorf("the request exceeded the maximum allowed payload size. You shouldn't encounter this under normal use")
	}

	if res.StatusCode == http.StatusUnprocessableEntity {
		return fmt.Errorf("the request data is invalid. This includes most of the base-specific validations. You will receive a detailed error message and code pointing to the exact issue")
	}

	if res.StatusCode == http.StatusInternalServerError {
		return fmt.Errorf("the server encountered an unexpected condition")
	}

	if res.StatusCode == http.StatusBadGateway {
		return fmt.Errorf("airtable's servers are restarting or an unexpected outage is in progress. You should generally not receive this error, and requests are safe to retry")
	}

	if res.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("the server could not process your request in time. The server could be temporarily unavailable, or it could have timed out processing your request. You should retry the request with backoffs")
	}

	if method == DELETE {
		return nil
	}

	if response != nil {
		return json.NewDecoder(res.Body).Decode(response)
	}

	return nil
}

// Attachment object may contain the following properties
type Attachment struct {
	ID         string `json:"id"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	URL        string `json:"url"`
	Filename   string `json:"filename"`
	Size       int    `json:"size"`
	Type       string `json:"type"`
	Thumbnails struct {
		Small struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"small"`
		Large struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"large"`
		Full struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"full"`
	} `json:"thumbnails"`
}
