package domain

type SearchItem struct {
	VodID      int
	VodName    string
	TypeName   string
	VodTime    string
	VodRemarks string
}

type SearchResult struct {
	SeriesIndex int
	DomainIndex int
	DomainURL   string
	Score       int
	Items       []SearchItem
	Err         error
}
