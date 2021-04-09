package tasks

import (
	"backend/processor/client"
	"backend/processor/lib"
	"backend/processor/model"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

type DeepScanTask struct {
	clients   client.Clients
	processor lib.Processor
}

func (task *DeepScanTask) Name() string {
	return "DEEP_SCAN"
}

func (task *DeepScanTask) Init(clients client.Clients) {
	task.clients = clients

	task.processor = lib.Processor{
		ApiClient:       task.clients.GetApiClient(),
		TaskProcessFunc: task.process,
		TaskName:        task.Name(),
		Parallelism:     150,
	}
}

func (task *DeepScanTask) Run() {
	task.processor.Start()

	log.Print(task.Name(), " task started processing")

	task.processor.Wait()

	log.Print(task.Name(), " task stopped processing")
}

func (task *DeepScanTask) process(item *model.PageRef) *model.PageRef {
	log.Tracef("page-ref received for download %s", item.Url)

	storeResult := task.clients.GetBackendStorageClient().Get(item)

	if !storeResult.Ok {
		item.Status = "FAILED"
	} else {
		item.Status = "FINISHED"
	}

	tokenizer := html.NewTokenizer(strings.NewReader(storeResult.Content))

	task.processTokens(tokenizer, item)

	return item
}

func (task *DeepScanTask) saveNewHref(href string, ref *model.PageRef) {
	newPageRef := makePageRef(href, ref)
	task.clients.GetApiClient().InsertPageRef(newPageRef)
}

func (task *DeepScanTask) processTokens(z *html.Tokenizer, pageRef *model.PageRef) {
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		token := z.Token()

		if token.Type == html.StartTagToken && token.Data == "a" {
			href := locateHref(token)

			if href != "" {
				href = fixHref(pageRef, href)

				if href == "" {
					continue
				}

				hrefParsed, _ := url.Parse(href)

				if hrefParsed == nil {
					continue
				}

				if !isHrefSuitable(href) {
					continue
				}

				if hrefParsed.Host != pageRef.WebsiteName &&
					!strings.Contains(hrefParsed.Host, pageRef.WebsiteName) {
					continue
				}

				task.saveNewHref(href, pageRef)
			}
		}
	}
}

func makePageRef(href string, parentPageRef *model.PageRef) *model.PageRef {
	id, _ := uuid.FromString(lib.NamedUUID([]byte(href)))

	return &model.PageRef{
		Id:          id,
		State:       "DOWNLOAD",
		Status:      "PENDING",
		Url:         href,
		WebsiteName: parentPageRef.WebsiteName,
	}
}

func isHrefSuitable(href string) bool {
	disallowedSuffixes := append([]string{}, ".jpg", ".png", ".gif")

	for _, disallowedSuffix := range disallowedSuffixes {
		if strings.HasSuffix(strings.ToLower(href), disallowedSuffix) {
			return false
		}
	}

	return true
}

func fixHref(ref *model.PageRef, href string) string {
	hrefUrl, err := url.Parse(href)

	if err != nil {
		log.Print(err)
		return ""
	}

	if hrefUrl.Scheme != "" && hrefUrl.Scheme != "http" && hrefUrl.Scheme != "https" {
		return ""
	}

	if !strings.HasPrefix(href, "http") {
		parentUrl, err := url.Parse(ref.Url)

		if err != nil {
			log.Print(err)
			return ""
		}

		if strings.HasPrefix(href, "//") {
			href = parentUrl.Scheme + ":" + href
		} else {
			// fixme, crappy logic
			baseUrl := parentUrl.Scheme + "://" + parentUrl.Hostname()

			if strings.HasPrefix(href, "/") {
				href = baseUrl + href
			} else if strings.Contains(parentUrl.Path, "/") {
				href = parentUrl.Path[0:strings.LastIndex(parentUrl.Path, "/")] + "/" + href
			} else {
				href = parentUrl.Path + "/" + href
			}

			href = baseUrl + hrefUrl.Path
			if !strings.HasPrefix(hrefUrl.Path, "/") {
				href = baseUrl + "/" + hrefUrl.Path
			}

			if len(hrefUrl.RawQuery) > 0 {
				href += "?" + hrefUrl.RawQuery
			}

			//log.Printf("href transforred from %s to %s via %s", oldHref, href, baseUrl)
		}
	}

	if !strings.HasPrefix(href, "http") {
		log.Print("defective url", href)
	}

	return href
}

func locateHref(token html.Token) string {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			return strings.TrimSpace(attr.Val)
		}
	}

	return ""
}
