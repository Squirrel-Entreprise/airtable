Simple pure Go Airtable API wrapper
================
[![GoDoc](https://godoc.org/github.com/Squirrel-Entreprise/airtable?status.svg)](https://pkg.go.dev/github.com/Squirrel-Entreprise/airtable)
![Go](https://github.com/Squirrel-Entreprise/airtable/workflows/Go/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/Squirrel-Entreprise/airtable)
[![codecov](https://codecov.io/gh/Squirrel-Entreprise/airtable/branch/main/graph/badge.svg)](https://codecov.io/gh/Squirrel-Entreprise/airtable)

## Installation

```go
    go get github.com/Squirrel-Entreprise/airtable
```

## Aitable API

Airtable uses simple token-based authentication. To generate or manage your API key, visit your [account](https://airtable.com/account) page.

## Getting started 

```go
a := airtable.New("xxx", "yyy")
```

### List table records 

```go
productsParameters := airtable.Parameters{
	Name:       "Products", // Name of the table
	MaxRecords: "100", // Max records to return
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
