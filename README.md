Very simple pure Go Airtable API wrapper
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

## Usage

```go
    package main

    import (
        "fmt"
        "github.com/Squirrel-Entreprise/airtable"
    )

    func main() {
        
        a := airtable.New("api_key_xxx", "id_base_yyy")

        productTable := airtable.Parameters{
            Name:            "Products",                         // Name of the table
            MaxRecords:      "100",                              // Max records to return
            View:            "Grid view",                        // View name
            FilterByFormula: fmt.Sprintf(`Name="%s"`, "Ananas"), // Filter by formula
            Fields: []string{ // Fields to return
                "Name",
                "Category",
            },
            Sort: []airtable.Sort{
                {
                    Field:     "Name",
                    Direction: airtable.Descending,
                },
            },
        }

        var products airtable.AirtableList

        if err := a.List(productTable, &products); err != nil {
            fmt.Println(err)
        }

        for _, p := range products.Records {
            fmt.Println(p.ID, p.Fields["Name"], p.Fields["Category"])
        }
    }
```

More examples can be found in `EXAMPLE.md`.