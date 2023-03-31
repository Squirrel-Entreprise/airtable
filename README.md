<!-- markdownlint-disable MD010 -->
# Golang Airtable API

[![GoDoc](https://godoc.org/github.com/Squirrel-Entreprise/airtable?status.svg)](https://pkg.go.dev/github.com/Squirrel-Entreprise/airtable)
![Go](https://github.com/Squirrel-Entreprise/airtable/workflows/Go/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/Squirrel-Entreprise/airtable)
[![codecov](https://codecov.io/gh/Squirrel-Entreprise/airtable/branch/main/graph/badge.svg)](https://codecov.io/gh/Squirrel-Entreprise/airtable)

A #golang package to access the [Airtable API](https://airtable.com/api).

## Table of contents

- [Golang Airtable API](#golang-airtable-api)
- [Table of contents](#table-of-contents)
  - [Installation](#installation)
  - [Installation](#installation)
  - [Aitable API](#aitable-api)
  - [Getting started](#getting-started)
    - [List table records](#list-table-records)
    - [Get table record](#get-table-record)
    - [Create table record](#create-table-record)
    - [Update table record](#update-table-record)
    - [Delete table record](#delete-table-record)

## Installation

```go
    go get github.com/Squirrel-Entreprise/airtable
```

## Aitable API

Airtable uses simple token-based authentication. To generate or manage your API key, visit your [account](https://airtable.com/account) page.

## Getting started

Initialize client

```go
a := airtable.New("xxx", "yyy")
```

### List table records

```go
productsParameters := airtable.Parameters{
	Name:       "Products", // Name of the table
	MaxRecords: "100", // Max records to return
    	PageSize:   "10",
	View:       "Grid view", // View name
	FilterByFormula: fmt.Sprintf(`Name="%s"`, "Apple"), // Filter by formula
	Fields: []string{ // Fields to return
		"Name",
		"Category",
	},
	Sort: []airtable.Sort{
		{
			Field:     "Category",
			Direction: airtable.Descending,
		},
	},
}

var products airtable.AirtableList

if err := a.List(productsParameters, &products); err != nil {
	fmt.Println(err)
}

for _, p := range products.Records {
	fmt.Println(p.ID, p.Fields["Name"], p.Fields["Category"])
}
```

#### Pagination

If you make a request that needs to return more records than fits in a _page_ (determined by Parameters.PageSize), the response AirtableList's Offset field will be non-empty.

For the small 5-record table below:

```none
| Name  |
| ----- |
| R-001 |
| R-002 |
| R-003 |
| R-004 |
| R-005 |
```

Getting a list with a page-size of 2:

```go
var (
	params = airtable.Parameters{
		Name:     "Products",
		PageSize: "2",
		Fields:   []string{"Name"},
	}
	resList airtable.AirtableList
)

a.List(params, &resList)
```

resList will have its Offset field set, in addition to its Records field:

```go
fmt.Println("offset:", resList.Offset)
for _, p := range resList.Records {
	fmt.Println("  name:", p.Fields["Name"])
}

// offset: itrUBoYIqbyPWClt7/rec15x8Y8iIFy0zLD
//   name: R-001
//   name: R-002
```

To get the next two products (R-003, R-004), add that offset into params before making the List call again:

```go
params.Offset = resList.Offset
resList.Offset = ""
a.List(params, &resList)
```

Printing resList like before:

```go
// offset: itr9dgLqRfeSslQ2G/rec1YiWuByyHHe2cM
//   name: R-003
//   name: R-004
```

Repeat to get the last product:

```go
params.Offset = resList.Offset
resList.Offset = ""
a.List(params, &resList)

// offset:
//   name: R-005
```

Note that offset is blank in the printout.  When the Airtable API has sent all the records it can, there will be no more pages, it omits the "offset" key in the JSON.  We need to be careful to always clear/reset resList.Offset to an empty string before making the call to List.  If we don't manually clear resList.Offset, it will keep **the old offset** and we won't get the signal from the API that there will be no more pages.

To make getting all records regardless of record count and page size, this client provides a ListPager.  Call your pagers Next method to get records back, at most page-size records at a time.  When all pages have been listed (no more records), Next returns an ErrEOL ("end of list").

Fetching the same five products two-at-a-time, like above, now looks like:

```go
pgr := NewListPager(a, params)
for {
	products, err := pgr.Next()
	if err != nil {
		if err == ErrEOL {
			break
		}
		log.Fatal(err)
	}

	fmt.Println("offset:", pgr.Offset())
	for _, p := range products {
		fmt.Println("  name:", p.Name)
	}
}
```

ListPager manages the offset logic, making sure you get all the records you expect with no fuss.

### Get table record

```go
product := airtable.AirtableItem{}
table := airtable.Parameters{Name: "Products"}
if err := a.Get(table, "recj2fwn8nSQhR9Gg", &product); err != nil {
	fmt.Println(err)
}

fmt.Println(product.ID, product.Fields["Name"], product.Fields["Category"])
```

### Create table record

```go
type porductPayload struct {
	Fields struct {
		Name     string  `json:"Name"`
		Category string  `json:"Category"`
		Price    float64 `json:"Price"`
	} `json:"fields"`
}

newProduct := porductPayload{}
newProduct.Fields.Name = "Framboise"
newProduct.Fields.Category = "Fruit"
newProduct.Fields.Price = 10.0

payload, err := json.Marshal(newProduct)
if err != nil {
	fmt.Println(err)
}

product := airtable.AirtableItem{}

table := airtable.Parameters{Name: "Products"}
if err := a.Create(table, payload, &product); err != nil {
	fmt.Println(err)
}

fmt.Println(product.ID, product.Fields["Name"], product.Fields["Price"])
```

### Update table record

```go
type porductPayload struct {
	Fields struct {
		Price float64 `json:"Price"`
	} `json:"fields"`
}

updateProduct := porductPayload{}
updateProduct.Fields.Price = 11.0

payload, err := json.Marshal(updateProduct)
if err != nil {
	fmt.Println(err)
}

product := airtable.AirtableItem{}

table := airtable.Parameters{Name: "Products"}
if err := a.Update(table, "recgnmCzr7u3jCB5w", payload, &product); err != nil {
	fmt.Println(err)
}

fmt.Println(product.ID, product.Fields["Name"], product.Fields["Price"])
```

### Delete table record

```go
table := airtable.Parameters{Name: "Products"}
if err := a.Delete(table, "recgnmCzr7u3jCB5w"); err != nil {
	fmt.Println(err)
}
```
