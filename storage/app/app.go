package app

import (
	"backend/storage/engine"
	"backend/storage/lib"
	"backend/storage/model"
	"crypto/tls"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"io"
	"net/http"
	"strings"
	"time"
)
import "github.com/gin-gonic/gin"

type App struct {
	Addr     string
	CertFile string
	KeyFile  string

	downloaderClient *http.Client
}

func (app *App) Run() {
	r := gin.New()

	app.routes(r)

	log.Info("Started\n")

	app.initDownloadClient()

	lib.Check(r.RunTLS(app.Addr, app.CertFile, app.KeyFile))
}

func (app *App) initDownloadClient() {
	app.downloaderClient = new(http.Client)
	app.downloaderClient.Timeout = time.Second * 100

	app.downloaderClient.Transport = &http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func (app *App) routes(r *gin.Engine) {
	r.POST("/api/1.0/get", app.get)
	r.POST("/api/1.0/store", app.store)
}

func (app *App) store(c *gin.Context) {
	pageRef := new(model.PageRef)

	err := c.BindJSON(pageRef)

	lib.Check(err)

	requestId := uuid.NewV4()
	lib.PageRefLogger(pageRef, "receive-store-request").Debugf("requestId: %s", requestId)

	storeResult, err := app.read(pageRef, false)

	lib.CheckWithPageRef(err, pageRef)

	if storeResult == nil {
		storeResult = new(model.StoreResult)
		storeResult.Ok = false
	}
	storeResult.Content = ""

	c.JSON(200, storeResult)
}

func (app *App) get(c *gin.Context) {
	pageRef := new(model.PageRef)

	err := c.BindJSON(pageRef)

	lib.Check(err)

	requestId := uuid.NewV4()
	lib.PageRefLogger(pageRef, "receive-get-request").Debugf("requestId: %s", requestId)

	storeResult, err := app.read(pageRef, true)

	lib.CheckWithPageRef(err, pageRef)

	c.JSON(200, storeResult)
}

func (app *App) read(pageRef *model.PageRef, require bool) (*model.StoreResult, error) {
	// check if download needed
	storage := app.getStorageBackend(pageRef)

	rec, err := storage.Get(pageRef.Id)

	if err != nil {
		return nil, err
	}

	forceDownload := pageRef.Tags != nil && contains(*pageRef.Tags, "force-download")
	lazyDownload := pageRef.Tags != nil && contains(*pageRef.Tags, "lazy-download")

	exists := rec != nil

	if len(rec) < 10000 {
		exists = false

		lib.PageRefLogger(pageRef, "delete-page-ref").Debugf("deleting")
		ok, err := storage.Delete(pageRef.Id)
		lib.PageRefLogger(pageRef, "delete-page-ref").Debugf("delete finished: %v", ok)

		if err != nil {
			return nil, err
		}
	}

	needsDownload := forceDownload || (!exists && (!lazyDownload || require))

	storeResult := new(model.StoreResult)
	var result []byte

	if needsDownload {

		downloadTry := 0
		for {
			downloadTry++

			lib.PageRefLogger(pageRef, "try-download").Tracef("downloading")
			result = app.download(pageRef.Url)
			lib.PageRefLogger(pageRef, "try-download").Tracef("download result: %d bytes", len(result))

			storeResult.Content = string(result)
			storeResult.Size = len(result)

			tryAllowed := storeResult.State == model.NO_CONTENT

			if tryAllowed {
				lib.PageRefLogger(pageRef, "try-download").Debugf("try count: %d", downloadTry)
			}

			if !tryAllowed {
				break // continue as no need to try
			}

			if tryAllowed && downloadTry > 10 {
				lib.PageRefLogger(pageRef, "try-download-fail").Warnf("no content")
				break
			}
		}

		checkStoreResult(storeResult, pageRef)

		if storeResult.Ok {
			if exists {
				_, err := storage.Delete(pageRef.Id)

				if err != nil {
					return nil, err
				}
			}

			lib.PageRefLogger(pageRef, "store-page-ref").Debugf("storing page ref")
			err = storage.Add(pageRef.Id, result)

			if err != nil {
				lib.PageRefLogger(pageRef, "store-page-ref-fail").Warnf("store page ref failed: %s", err)
				return nil, err
			}
		} else {
			log.Warnf("page-ref failed to download %s %d ", pageRef.Url, downloadTry)
		}

		return storeResult, err
	} else {
		if rec != nil {
			storeResult.State = model.ALREADY_DOWNLOADED
			storeResult.Content = string(rec)
			storeResult.Size = len(rec)
			checkStoreResult(storeResult, pageRef)
		} else {
			storeResult.State = model.SKIPPED
		}

		storeResult.Ok = true
		return storeResult, nil
	}
}

func checkStoreResult(result *model.StoreResult, pageRef *model.PageRef) {
	if result.Size < 10000 {
		if strings.Contains(result.Content, "DDoS protection") {
			log.Warn("cloudflare protection: " + pageRef.Url)
			result.State = model.CLOUDFLARE_DDOS_PROTECTION
		}

		if len(result.Content) > 0 {
			log.Warnf("low content size %s SIZE : %d", pageRef.Url, len(result.Content))
			log.Debugf("low content size %s CONTENT : %s", pageRef.Url, result.Content)
			result.State = model.LOW_CONTENT_SIZE
		} else if len(result.Content) == 0 {
			log.Warnf("no content for %s", pageRef.Url) // try again
			result.State = model.NO_CONTENT
		}

		result.Ok = false
	} else {
		result.Ok = true
		result.State = model.DOWNLOADED
		log.Debugf("page-ref downloaded %s", pageRef.Url)
	}
}

func (app *App) download(url string) []byte {
	resp, err := app.downloaderClient.Get("https://tisserv.net:8234/get?url=" + url)

	if err != nil {
		log.Print(err)
		return nil
	}

	defer resp.Body.Close()

	data := new(strings.Builder)
	_, err = io.Copy(data, resp.Body)

	if err != nil {
		log.Print(err)
		return nil
	}

	content := data.String()

	if strings.Contains(content, "DDoS protection by") {
		log.Printf("DDoS protection by page received: %s", url)
	}

	return []byte(content)
}

func (app *App) getStorageBackend(ref *model.PageRef) engine.StorageEngineBackend {
	if contains(*ref.Tags, "mongo-old") {
		return engine.GetEngineBackendByClassName("mongo-old")
	} else {
		return engine.GetEngineBackendByClassName("mongo-store")
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
