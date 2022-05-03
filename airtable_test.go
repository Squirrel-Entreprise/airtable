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

func TestSetXAPIKey(t *testing.T) {
	a := New("xxx", "yyy")

	a.SetXAPIKey("xxx")
}

func TestListBases(t *testing.T) {
	a := New("xxx", "yyy")
	Client = &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/v0/meta/bases" {
				t.Errorf("Expected to request '/v0/meta/bases', got: %s", req.URL.Path)
			}

			responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{
					"bases": [
					{
						"id": "appY3WxIBCdKPDdIa",
						"name": "Apartment Hunting",
						"permissionLevel": "create"
					},
					{
						"id": "appSW9R5uCNmRmfl6",
						"name": "Project Tracker",
						"permissionLevel": "edit"
					}
				]
			}`)))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       responseBody,
			}, nil
		},
	}
	_, err := a.ListBases()
	if err != nil {
		t.Errorf("list bases should not return error, got %s", err)
	}

}

func TestBaseSchema(t *testing.T) {
	a := New("xxx", "yyy")
	Client = &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/v0/meta/bases/xxx/tables" {
				t.Errorf("Expected to request '/v0/meta/bases/xxx/tables', got: %s", req.URL.Path)
			}

			responseBody := ioutil.NopCloser(bytes.NewReader([]byte(`{
				"tables": [
				  {
					"id": "tbltp8DGLhqbUmjK1",
					"name": "Apartments",
					"description": "Apartments to track.",
					"primaryFieldId": "fld1VnoyuotSTyxW1",
					"fields": [
					  {
						"id": "fld1VnoyuotSTyxW1",
						"name": "Name",
						"description": "Name of the apartment",
						"type": "singleLineText"
					  },
					  {
						"id": "fldoaIqdn5szURHpw",
						"name": "Pictures",
						"type": "multipleAttachment"
					  },
					  {
						"id": "fldumZe00w09RYTW6",
						"name": "District",
						"type": "multipleRecordLinks",
						"options": {
						  "isReversed": false,
						  "inverseLinkFieldId": "fldWnCJlo2z6ttT8Y",
						  "linkedTableId": "tblK6MZHez0ZvBChZ",
						  "prefersSingleRecordLink": true
						}
					  }
					],
					"views": [
					  {
						"id": "viwQpsuEDqHFqegkp",
						"name": "Grid view",
						"type": "grid"
					  }
					]
				  },
				  {
					"id": "tblK6MZHez0ZvBChZ",
					"name": "Districts",
					"primaryFieldId": "fldEVzvQOoULO38yl",
					"fields": [
					  {
						"id": "fldEVzvQOoULO38yl",
						"name": "Name",
						"type": "singleLineText"
					  },
					  {
						"id": "fldWnCJlo2z6ttT8Y",
						"name": "Apartments",
						"description": "Apartments that belong to this district",
						"type": "multipleRecordLinks",
						"options": {
						  "isReversed": false,
						  "inverseLinkFieldId": "fldumZe00w09RYTW6",
						  "linkedTableId": "tbltp8DGLhqbUmjK1",
						  "prefersSingleRecordLink": false
						}
					  }
					],
					"views": [
					  {
						"id": "viwi3KXvrKug2mIBS",
						"name": "Grid view",
						"type": "grid"
					  }
					]
				  }
				]
			  }`)))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       responseBody,
			}, nil
		},
	}
	_, err := a.BaseSchema("xxx")
	if err != nil {
		t.Errorf("base schema should not return error, got %s", err)
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

		if err := a.List(Parameters{}, nil); err == nil {
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

		param := Parameters{
			Name:   "test",
			Fields: []string{"ok", "ok2"},
		}
		var r AirtableList
		if err := a.List(param, &r); err != nil {
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

		param := Parameters{
			Name: "test",
			Sort: []Sort{
				{
					Field:     "ok",
					Direction: Descending,
				},
			},
		}
		var r AirtableList
		if err := a.List(param, &r); err != nil {
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

		param := Parameters{
			Name:            "test",
			FilterByFormula: "ok = 'ok'",
		}
		var r AirtableList
		if err := a.List(param, &r); err != nil {
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
		if err := a.Get(Parameters{}, "", &r); err == nil {
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
		if err := a.Get(Parameters{Name: "test"}, "id", &r); err != nil {
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
		if err := a.Create(Parameters{}, nil, &r); err == nil {
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
		if err := a.Create(Parameters{Name: "test"}, nil, &r); err != nil {
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
		if err := a.Update(Parameters{}, "id", nil, &r); err == nil {
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
		if err := a.Update(Parameters{Name: "test"}, "id", nil, &r); err != nil {
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

		if err := a.Delete(Parameters{}, ""); err == nil {
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

		if err := a.Delete(Parameters{Name: "test"}, "id"); err != nil {
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

	t.Run("x-airtable-client-secret", func(t *testing.T) {
		a.SetXAPIKey("xxx")
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
