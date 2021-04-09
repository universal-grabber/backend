package model

type StoreResult struct {
	Content string     `json:"content"`
	Size    int        `json:"size"`
	Ok      bool       `json:"ok"`
	State   StoreState `json:"state"`
}

type StoreState string

const (
	CLOUDFLARE_DDOS_PROTECTION StoreState = "CLOUDFLARE_DDOS_PROTECTION"
	LOW_CONTENT_SIZE           StoreState = "LOW_CONTENT_SIZE"
	NO_CONTENT                 StoreState = "NO_CONTENT"
	ALREADY_DOWNLOADED         StoreState = "ALREADY_DOWNLOADED"
	DOWNLOADED                 StoreState = "DOWNLOADED"
	SKIPPED                    StoreState = "SKIPPED"
)
