package model

type WebSite struct {
	Id          string `bson:"_id" json:"id"`
	Name        string `json:"name"`
	Url         string `json:"url"`
	BatchDelay  int    `json:"batch_delay"`
	BatchSize   int    `json:"batch_size"`
	TagMatch    []TagMatcher
	EntryPoints []string
	Schedule    []WebsiteSchedule
}

type TagMatcher struct {
	Tags     []string
	Pattern  string
	Patterns []string
}

type WebsiteSchedule struct {
	Match     *SearchPageRef
	Frequency string
	ToState   string
	ToStatus  string
}
