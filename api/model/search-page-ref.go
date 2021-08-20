package model

type SearchPageRef struct {
	WebsiteName string
	State       string
	Status      string
	FairSearch  bool // search in each website equally
	Page        int
	PageSize    int
	Tags        []string
	ProcessAll  bool
	Provision   bool
}
