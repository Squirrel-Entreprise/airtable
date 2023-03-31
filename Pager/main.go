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

var ErrEOL = errors.New("no more pages in list")

// A ListPager iterates the pages of a list by tracking
// page offsets.
type ListPager struct {
	at *airtable.Airtable
	pm airtable.Parameters
	lo airtable.AirtableList // "list object"

	done bool // track last page
}

func NewListPager(at *airtable.Airtable, pm airtable.Parameters) *ListPager {
	return &ListPager{at, pm, airtable.AirtableList{}, false}
}

// Next fetches the next available page from at.List(...) and
// returns the slice of records.
// If the list has been exhausted (no more pages)
// Next returns nil, ErrEOL.
func (p *ListPager) Next() (records []airtable.AirtableItem, err error) {
	if p.done {
		return nil, ErrEOL
	}

	// set params with last offset, clear lo's offset
	p.pm.Offset, p.lo.Offset = p.lo.Offset, ""
	if err := p.at.List(p.pm, &p.lo); err != nil {
		return nil, err
	}

	p.done = p.lo.Offset == ""

	return p.lo.Records, nil
}

func (p *ListPager) Offset() string {
	return p.lo.Offset
}

func main() {
	b, _ := os.ReadFile(".secret.txt")
	apiKey := string(b)

	at := airtable.New(apiKey, appID, false)
	pm := airtable.Parameters{
		Name:       tblID,
		MaxRecords: "10",
		PageSize:   "3",
	}

	pgr := NewListPager(at, pm)
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
