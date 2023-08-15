package pagination

import "math"

type Pagination struct {
	NumPages           int
	HasPrev, HasNext   bool
	PrevPage, NextPage int
	ItemsPerPage       int
	CurrentPage        int
	NumItems           int
	Start, End         int
}

func Calculate(currentPage, itemsPerPage, numItems int) (p Pagination) {
	p = Pagination{}

	p.CurrentPage = currentPage
	p.ItemsPerPage = itemsPerPage
	p.NumItems = numItems

	// calc start + end
	p.Start = (currentPage - 1) * p.ItemsPerPage

	if p.Start > p.NumItems {
		p.Start = p.NumItems
	}

	p.End = p.Start + p.ItemsPerPage
	if p.End > p.NumItems {
		p.End = p.NumItems
	}

	// calc number of pages
	d := float64(p.NumItems) / float64(p.ItemsPerPage)
	p.NumPages = int(math.Ceil(d))

	// HasPrev, HasNext?
	p.HasPrev = p.CurrentPage > 1
	p.HasNext = p.CurrentPage < p.NumPages

	// calculate prev + next pages
	if p.HasPrev {
		p.PrevPage = p.CurrentPage - 1
	}
	if p.HasNext {
		p.NextPage = p.CurrentPage + 1
	}

	return
}
