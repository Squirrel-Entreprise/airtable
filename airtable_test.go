package airtable

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

func TestList(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("empty table name", func(t *testing.T) {
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

		if err := a.List(Table{}, nil); err == nil {
			t.Errorf("table name is required, got %s", err)
		}
	})

	t.Run("fields", func(t *testing.T) {
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

		table := Table{
			Name:   "test",
			Fields: []string{"ok", "ok2"},
		}
		var r AirtableList
		if err := a.List(table, &r); err != nil {
			t.Errorf("list should not return error, got %s", err)
		}
	})

	t.Run("sort", func(t *testing.T) {
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

		table := Table{
			Name: "test",
			Sort: []Sort{
				{
					Field:     "ok",
					Direction: Descending,
				},
			},
		}
		var r AirtableList
		if err := a.List(table, &r); err != nil {
			t.Errorf("list should not return error, got %s", err)
		}
	})

	t.Run("filter by formula", func(t *testing.T) {
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

		table := Table{
			Name:            "test",
			FilterByFormula: "ok = 'ok'",
		}
		var r AirtableList
		if err := a.List(table, &r); err != nil {
			t.Errorf("list should not return error, got %s", err)
		}
	})
}

func TestGet(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("empty table name", func(t *testing.T) {
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

		var r AirtableItem
		if err := a.Get(Table{}, "", &r); err == nil {
			t.Errorf("table name is required, got %s", err)
		}
	})

	t.Run("get", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/id" {
					t.Errorf("Expected to request '/v0/yyy/test/id', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		var r AirtableItem
		if err := a.Get(Table{Name: "test"}, "id", &r); err != nil {
			t.Errorf("get should not return error, got %s", err)
		}
	})
}

func TestCreate(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("empty table name", func(t *testing.T) {
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

		var r AirtableItem
		if err := a.Create(Table{}, nil, &r); err == nil {
			t.Errorf("table name is required, got %s", err)
		}
	})

	t.Run("create", func(t *testing.T) {
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

		var r AirtableItem
		if err := a.Create(Table{Name: "test"}, nil, &r); err != nil {
			t.Errorf("create should not return error, got %s", err)
		}
	})
}

func TestUpdate(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("empty table name", func(t *testing.T) {
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

		var r AirtableItem
		if err := a.Update(Table{}, "id", nil, &r); err == nil {
			t.Errorf("table name is required, got %s", err)
		}
	})

	t.Run("update", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/id" {
					t.Errorf("Expected to request '/v0/yyy/test/id', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		var r AirtableItem
		if err := a.Update(Table{Name: "test"}, "id", nil, &r); err != nil {
			t.Errorf("update should not return error, got %s", err)
		}
	})
}

func TestDelete(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("empty table name", func(t *testing.T) {
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

		if err := a.Delete(Table{}, ""); err == nil {
			t.Errorf("table name is required, got %s", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/yyy/test/id" {
					t.Errorf("Expected to request '/v0/yyy/test/id', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		if err := a.Delete(Table{Name: "test"}, "id"); err != nil {
			t.Errorf("delete should not return error, got %s", err)
		}
	})
}

func TestCall(t *testing.T) {
	a := New("xxx", "yyy")

	t.Run("client_do", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, fmt.Errorf("client_do")
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("call should return error, got %s", err)
		}
	})

	t.Run("bad_request", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("payment_required", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusPaymentRequired,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("request_entity_too_large", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusRequestEntityTooLarge,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("unprocessable_entity", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusUnprocessableEntity,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("internal_server_error", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("bad_gateway", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusBadGateway,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("service_unavailable", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("too_many_requests", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       responseBody,
				}, nil
			},
		}
		if err := a.call(GET, &url.URL{}, nil, nil); err == nil {
			t.Errorf("Expected to return error, got %s", err)
		}
	})

	t.Run("response_nil", func(t *testing.T) {
		Client = &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/v0/" {
					t.Errorf("Expected to request '/v0/', got: %s", req.URL.Path)
				}

				responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{"value":"fixed"}`)))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody,
				}, nil
			},
		}

		if err := a.call(GET, &url.URL{}, nil, nil); err != nil {
			t.Errorf("Expected to return nil, got %s", err)
		}
	})
}
