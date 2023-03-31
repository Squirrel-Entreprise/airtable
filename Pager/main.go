package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Squirrel-Entreprise/airtable"
)

const (
	appID = "app3n84SwUnqPsvgC"
	tblID = "Table_1"
)

var (
	apiKey string

	a  *airtable.Airtable
	p  airtable.Parameters
	lo airtable.AirtableList
	o  string
)

const initOffset = "InitialOffset"

func init() {
	b, _ := os.ReadFile(".secret.txt")
	apiKey = string(b)

	a = airtable.New(apiKey, appID, false)
	p = airtable.Parameters{
		Name:       tblID,
		MaxRecords: "10",
		PageSize:   "3",
	}
	o = initOffset
}

var ErrEOL = errors.New("no more pages in list")

func next() error {
	switch o {
	case "":
		return ErrEOL
	case initOffset:
		o = ""
	}

	p.Offset = o
	lo.Offset = "" // lo.Offset must be reset everytime
	if err := a.List(p, &lo); err != nil {
		return err
	}
	o = lo.Offset

	return nil
}

func main() {
	for {
		if err := next(); err != nil {
			if err == ErrEOL {
				break
			}
			log.Fatal(err)
		}
		printLO()
	}
}

func printLO() {
	fmt.Println("offset:", lo.Offset)
	for _, record := range lo.Records {
		fmt.Println("    recID:", record.ID)
	}
}
