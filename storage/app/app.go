package app

import (
	"backend/common"
	"backend/gen/proto/base"
	pb "backend/gen/proto/service/storage"
	"backend/processor/lib"
	"backend/storage/engine"
	"context"
	"crypto/tls"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type App struct {
	Addr     string
	CertFile string
	KeyFile  string

	downloaderClient *http.Client

	pb.UnimplementedStorageServiceServer
}

func (app *App) Run() {
	log.Info("Started\n")

	app.initDownloadClient()

	app.runGrpc()
}

func (app *App) runGrpc() {
	lis, err := net.Listen("tcp", "0.0.0.0:6565")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterStorageServiceServer(s, app)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func (app *App) initDownloadClient() {
	app.downloaderClient = new(http.Client)
	app.downloaderClient.Timeout = time.Second * 100

	app.downloaderClient.Transport = &http2.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func (app *App) Get(ctx context.Context, pageRef *base.PageRef) (*pb.StoreResult, error) {
	requestId := uuid.NewV4()
	common.PageRefLogger(pageRef, "receive-get-request").Debugf("requestId: %s", requestId)

	storeResult, err := app.read(pageRef, true)

	common.CheckWithPageRef(err, pageRef)

	return storeResult, nil
}
func (app *App) Store(ctx context.Context, pageRef *base.PageRef) (*pb.StoreResult, error) {
	requestId := uuid.NewV4()
	common.PageRefLogger(pageRef, "receive-store-request").Debugf("requestId: %s", requestId)

	storeResult, err := app.read(pageRef, false)

	common.CheckWithPageRef(err, pageRef)

	if storeResult == nil {
		storeResult = new(pb.StoreResult)
		storeResult.Ok = false
	}
	storeResult.Content = ""

	return storeResult, nil
}

func (app *App) read(pageRef *base.PageRef, require bool) (*pb.StoreResult, error) {
	// check if download needed
	storage := app.getStorageBackend(pageRef)

	rec, err := storage.Get(toUUID(pageRef.GetId()))

	if err != nil {
		return nil, err
	}

	forceDownload := pageRef.Tags != nil && contains(pageRef.Tags, "force-download")
	lazyDownload := pageRef.Tags != nil && contains(pageRef.Tags, "lazy-download")

	exists := rec != nil

	//if len(rec) < 10000 {
	//	exists = false
	//
	//	common.PageRefLogger(pageRef, "delete-page-ref").Debugf("deleting")
	//	ok, err := storage.Delete(toUUID(pageRef.Id))
	//	common.PageRefLogger(pageRef, "delete-page-ref").Debugf("delete finished: %v", ok)
	//
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	needsDownload := forceDownload || (!exists && (!lazyDownload || require))

	storeResult := new(pb.StoreResult)
	var result []byte

	if needsDownload {

		downloadTry := 0
		for {
			downloadTry++

			common.PageRefLogger(pageRef, "try-download").Tracef("downloading")
			result = app.download(pageRef.Url)
			common.PageRefLogger(pageRef, "try-download").Tracef("download result: %d bytes", len(result))

			storeResult.Content = string(result)
			storeResult.Size = int32(len(result))

			tryAllowed := storeResult.State == pb.State_NO_CONTENT

			if tryAllowed {
				common.PageRefLogger(pageRef, "try-download").Debugf("try count: %d", downloadTry)
			}

			if !tryAllowed {
				break // continue as no need to try
			}

			if tryAllowed && downloadTry > 10 {
				common.PageRefLogger(pageRef, "try-download-fail").Warnf("no content")
				break
			}
		}

		checkStoreResult(storeResult, pageRef)

		if storeResult.Ok {
			if exists {
				_, err := storage.Delete(toUUID(pageRef.GetId()))

				if err != nil {
					return nil, err
				}
			}

			common.PageRefLogger(pageRef, "store-page-ref").Debugf("storing page ref")
			err = storage.Add(toUUID(pageRef.GetId()), result)

			if err != nil {
				common.PageRefLogger(pageRef, "store-page-ref-fail").Warnf("store page ref failed: %s", err)
				return nil, err
			}
		} else {
			log.Warnf("page-ref failed to download %s %d ", pageRef.Url, downloadTry)
		}

		return storeResult, err
	} else {
		log.Info("page-ref-download-skipped %s %s %s %s", pageRef.Url, forceDownload, exists, lazyDownload, require)
		if rec != nil {
			storeResult.State = pb.State_ALREADY_DOWNLOADED
			storeResult.Content = string(rec)
			storeResult.Size = int32(len(rec))
			checkStoreResult(storeResult, pageRef)
		} else {
			storeResult.State = pb.State_SKIPPED
		}

		storeResult.Ok = true
		return storeResult, nil
	}
}

func toUUID(id string) uuid.UUID {
	res, err := uuid.FromString(id)

	lib.Check(err)

	return res
}

func checkStoreResult(result *pb.StoreResult, pageRef *base.PageRef) {
	//if result.Size < 10000 {
	//	if strings.Contains(result.Content, "DDoS protection") {
	//		log.Warn("cloudflare protection: " + pageRef.Url)
	//		result.State = pb.State_CLOUDFLARE_DDOS_PROTECTION
	//	}
	//
	//	if len(result.Content) > 0 {
	//		log.Warnf("low content size %s SIZE : %d", pageRef.Url, len(result.Content))
	//		log.Debugf("low content size %s CONTENT : %s", pageRef.Url, result.Content)
	//		result.State = pb.State_LOW_CONTENT_SIZE
	//	} else if len(result.Content) == 0 {
	//		log.Warnf("no content for %s", pageRef.Url) // try again
	//		result.State = pb.State_NO_CONTENT
	//	}
	//
	//	result.Ok = false
	//} else {
	result.Ok = true
	result.State = pb.State_DOWNLOADED
	log.Debugf("page-ref downloaded %s", pageRef.Url)
	//}
}

func (app *App) download(url string) []byte {
	downloadUrl := "https://ug.tisserv.net:8234/get-clean?noProxy=true&url=" + url
	log.Print("Download task: ", downloadUrl)
	resp, err := app.downloaderClient.Get(downloadUrl)

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

func (app *App) getStorageBackend(ref *base.PageRef) engine.StorageEngineBackend {
	if contains(ref.Tags, "mongo-old") {
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
