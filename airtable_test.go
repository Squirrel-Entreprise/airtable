package airtable

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestNew(t *testing.T) {
	a := New("xxx", "yyy")

	if a.apiKey != "xxx" {
		t.Errorf("apiKey should be xxx, got %s", a.apiKey)
	}

	if a.base != "yyy" {
		t.Errorf("base should be yyy, got %s", a.base)
	}
}

func TestCall(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("empty table name", func(t *testing.T) {
		if err := a.call(GET, Table{}, nil, nil, nil); err == nil {
			t.Errorf("table name is required, got %s", err)
		}
	})

	t.Run("client_do", func(t *testing.T) {
		id := "123"
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, fmt.Errorf("client_do")
			},
		}
		if err := a.call(GET, Table{Name: "test"}, &id, nil, nil); err == nil {
			t.Errorf("call should return error, got %s", err)
		}
	})

	t.Run("list path", func(t *testing.T) {
		if err := a.call(GET, Table{Name: "test"}, nil, nil, nil); err == nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("get path || delete path || update path", func(t *testing.T) {
		id := "123"
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/"+id {
					t.Errorf("Expected to request '/v0/yyy/test/"+id+"', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, Table{Name: "test"}, &id, nil, nil); err != nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("create path", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err != nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("delete path", func(t *testing.T) {
		id := "123"
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/"+id {
					t.Errorf("Expected to request '/v0/yyy/test/"+id+"', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(DELETE, Table{Name: "test"}, &id, nil, nil); err != nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("bad_request", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("payment_required", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusPaymentRequired,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("request_entity_too_large", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusRequestEntityTooLarge,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("unprocessable_entity", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("internal_server_error", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("bad_gateway", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("service_unavailable", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("too_many_requests", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("response_nil", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		var response Attachment
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), &response); err != nil {
			t.Errorf("Expected to return nil, got %s", err)
		}
	})
}

func TestList(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("not empty interface responce", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.List(Table{Name: "test"}, &r); err != nil {
			t.Errorf("list should not return error, got %s", err)
		}
	})
}

func TestGet(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("not empty interface responce", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/" {
					t.Errorf("Expected to request '/v0/yyy/test/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Get(Table{Name: "test"}, "", &r); err != nil {
			t.Errorf("get should not return error, got %s", err)
		}
	})
}

func TestCreate(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("422 unprocessable entity", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       responseBody,
				}, nil
			},
		}
		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Create(Table{Name: "test"}, []byte(`ok`), &r); err != nil {
			if !strings.Contains(err.Error(), "request data is invalid") {
				t.Errorf("create should return 'request data is invalid', got %s", err)
			}
		}
	})

	t.Run("not empty interface responce", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test" {
					t.Errorf("Expected to request '/v0/yyy/test', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Create(Table{Name: "test"}, []byte(`{"fields": {"Name": "Ail"}}`), &r); err != nil {
			t.Errorf("create should not return error, got %s", err)
		}
	})
}

func TestUpdate(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("422 unprocessable entity", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/" {
					t.Errorf("Expected to request '/v0/yyy/test/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       responseBody,
				}, nil
			},
		}

		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Update(Table{Name: "test"}, "", []byte(`ok`), &r); err != nil {
			if !strings.Contains(err.Error(), "request data is invalid") {
				t.Errorf("create should return 'request data is invalid', got %s", err)
			}
		}
	})

	t.Run("not empty interface responce", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/" {
					t.Errorf("Expected to request '/v0/yyy/test/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Update(Table{Name: "test"}, "", []byte(`{"fields": {"Name": "Ail"}}`), &r); err != nil {
			t.Errorf("update should not return error, got %s", err)
		}
	})
}

func TestDelete(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/" {
					t.Errorf("Expected to request '/v0/yyy/test/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.Delete(Table{Name: "test"}, ""); err != nil {
			t.Errorf("delete should not return error, got %s", err)
		}
	})
}
