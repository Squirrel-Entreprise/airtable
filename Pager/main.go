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

	// a  *airtable.Airtable
	// p  airtable.Parameters
	// lo airtable.AirtableList
	// o  string
)

const initOffset = "InitialOffset" // Need a non-empty value for first iteration

func init() {
	b, _ := os.ReadFile(".secret.txt")
	apiKey = string(b)

	// a = airtable.New(apiKey, appID, false)
	// p = airtable.Parameters{
	// 	Name:       tblID,
	// 	MaxRecords: "10",
	// 	PageSize:   "3",
	// }
	// o = initOffset
}

type ListPager struct {
	at *airtable.Airtable
	pm airtable.Parameters
	lo airtable.AirtableList

	o string
}

func NewListPager(at *airtable.Airtable, pm airtable.Parameters, lo airtable.AirtableList) *ListPager {
	return &ListPager{at, pm, lo, initOffset}
}

var ErrEOL = errors.New("no more pages in list")

func (lp *ListPager) Next() (records []airtable.AirtableItem, err error) {
	switch lp.o {
	case "":
		return nil, ErrEOL
	case initOffset:
		lp.o = ""
	}

	lp.pm.Offset = lp.o
	lp.lo.Offset = "" // lo.Offset must be reset everytime
	if err := lp.at.List(lp.pm, &lp.lo); err != nil {
		return nil, err
	}
	lp.o = lp.lo.Offset

	return lp.lo.Records, nil
}

func (p *ListPager) Offset() string {
	return p.o
}

/*
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
*/

func main() {
	var (
		a = airtable.New(apiKey, appID, false)
		p = airtable.Parameters{
			Name: tblID,
			// MaxRecords: "10",
			// PageSize:   "3",
		}
		l airtable.AirtableList
	)

	pgr := NewListPager(a, p, l)
	for {
		records, err := pgr.Next()
		if err != nil {
			if err == ErrEOL {
				break
			}
			log.Fatal(err)
		}
		fmt.Println("o:", pgr.Offset())
		for _, record := range records {
			fmt.Println("  r:", record.ID)
		}
	}
}

/* using next()
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
	fmt.Println("o:", lo.Offset)
	for _, record := range lo.Records {
		fmt.Println("  r:", record.ID)
	}
}
*/
