package airtable

import (
	"strings"
	"testing"
)

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
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("list path", func(t *testing.T) {
		if err := a.call(GET, Table{Name: "test"}, nil, nil, nil); err == nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("get path || delete path || update path", func(t *testing.T) {
		id := "123"
		if err := a.call(GET, Table{Name: "test"}, &id, nil, nil); err == nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("create path", func(t *testing.T) {
		if err := a.call(POST, Table{Name: "test"}, nil, []byte(`ok`), nil); err == nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})

	t.Run("delete path", func(t *testing.T) {
		id := "123"
		if err := a.call(DELETE, Table{Name: "test"}, &id, nil, nil); err != nil {
			t.Errorf("call should not return error, got %s", err)
		}
	})
}

func TestList(t *testing.T) {
	a := New("xxx", "yyy")
	t.Run("empty interface responce", func(t *testing.T) {
		if err := a.List(Table{Name: "test"}, nil); err == nil {
			t.Errorf("list should not return error, got %s", err)
		}
	})

	t.Run("not empty interface responce", func(t *testing.T) {
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
	t.Run("empty interface responce", func(t *testing.T) {
		if err := a.Get(Table{Name: "test"}, "", nil); err == nil {
			t.Errorf("get should not return error, got %s", err)
		}
	})

	t.Run("not empty interface responce", func(t *testing.T) {
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
	t.Run("empty interface responce", func(t *testing.T) {
		if err := a.Create(Table{Name: "test"}, []byte(`ok`), nil); err == nil {
			t.Errorf("create should not return error, got %s", err)
		}
	})

	t.Run("422 unprocessable entity", func(t *testing.T) {
		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Create(Table{Name: "test"}, []byte(`ok`), &r); err != nil {
			if !strings.Contains(err.Error(), "422") {
				t.Errorf("create should return 422, got %s", err)
			}
		}
	})

	t.Run("not empty interface responce", func(t *testing.T) {
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
	t.Run("empty interface responce", func(t *testing.T) {
		if err := a.Update(Table{Name: "test"}, "", []byte(`ok`), nil); err == nil {
			t.Errorf("update should not return error, got %s", err)
		}
	})

	t.Run("422 unprocessable entity", func(t *testing.T) {
		type resp struct {
			Records interface{} `json:"records"`
			Offset  string      `json:"offset"`
		}
		var r resp
		if err := a.Update(Table{Name: "test"}, "", []byte(`ok`), &r); err != nil {
			if !strings.Contains(err.Error(), "422") {
				t.Errorf("update should return 422, got %s", err)
			}
		}
	})

	t.Run("not empty interface responce", func(t *testing.T) {
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
	t.Run("empty table name", func(t *testing.T) {
		if err := a.Delete(Table{Name: ""}, ""); err == nil {
			t.Errorf("delete should not return error, got %s", err)
		}
	})

	t.Run("", func(t *testing.T) {
		if err := a.Delete(Table{Name: "test"}, ""); err != nil {
			t.Errorf("delete should not return error, got %s", err)
		}
	})
}
