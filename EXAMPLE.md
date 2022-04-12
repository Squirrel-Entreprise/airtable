Example
================

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Squirrel-Entreprise/airtable"
)

func main() {
	a := airtable.New("xxx", "yyy")

	productTable := airtable.Table{
		Name:       "Products",
		MaxRecords: "100",
		View:       "Grid view",
	}

	type productItemAirtable struct {
		ID          string    `json:"id"`
		CreatedTime time.Time `json:"createdTime"`
		Fields      struct {
			Name     string                `json:"Name"`
			Cover    []airtable.Attachment `json:"cover"`
			Category string                `json:"Category"`
			Price    float64               `json:"Price"`
			Carts    []string              `json:"Carts"`
		} `json:"fields"`
	}

	// List products
	type productsListAirtable struct {
		Records []productItemAirtable `json:"records"`
		Offset  string                `json:"offset"`
	}

	products := productsListAirtable{}

	if err := a.List(productTable, &products); err != nil {
		fmt.Println(err)
	}

	for _, p := range products.Records {
		fmt.Println(p.ID, p.Fields.Name, p.Fields.Price)
	}

	// Get product
	product := productItemAirtable{}

	if err := a.Get(productTable, products.Records[0].ID, &product); err != nil {
		fmt.Println(err)
	}

	fmt.Println(product.ID, product.Fields.Name, product.Fields.Price)

	type porductPayload struct {
		Fields struct {
			Name  string `json:"Name"`
			Cover []struct {
				URL string `json:"url"`
			} `json:"cover"`
			Category string   `json:"Category"`
			Price    float64  `json:"Price"`
			Carts    []string `json:"Carts"`
		} `json:"fields"`
	}

	// Create product
	newProduct := porductPayload{}
	newProduct.Fields.Name = "New product"
	newProduct.Fields.Price = 10.0
	newProduct.Fields.Category = "Fruit"

	payload, err := json.Marshal(newProduct)
	if err != nil {
		fmt.Println(err)
	}

	if err := a.Create(productTable, payload, &product); err != nil {
		fmt.Println(err)
	}

	fmt.Println(product.ID, product.Fields.Name, product.Fields.Price)

	// Update product
	updateProduct := porductPayload{}
	updateProduct.Fields.Name = "New product Updated"
	updateProduct.Fields.Price = 10.0
	updateProduct.Fields.Category = "Légume"

	payloadUpdate, err := json.Marshal(updateProduct)
	if err != nil {
		fmt.Println(err)
	}

	if err := a.Update(productTable, product.ID, payloadUpdate, &product); err != nil {
		fmt.Println(err)
	}

	fmt.Println(product.ID, product.Fields.Name, product.Fields.Price)

	// Delete product
	if err := a.Delete(productTable, product.ID); err != nil {
		fmt.Println(err)
	}
}

```