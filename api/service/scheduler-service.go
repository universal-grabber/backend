package service

import (
	"backend/api/const"
	"backend/api/helper"
	"backend/api/model"
	"backend/api/util"
	"context"
	"github.com/robfig/cron"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/url"
	"regexp"
	"strings"
)

type SchedulerServiceImpl struct {
	tagMatchers map[string][]model.TagMatcher
	regexCache  map[string]*regexp.Regexp

	service *PageRefService
}

func (s *SchedulerServiceImpl) Run() {
	s.tagMatchers = make(map[string][]model.TagMatcher)
	s.regexCache = make(map[string]*regexp.Regexp)
	s.service = new(PageRefService)

	c := cron.New()

	go s.reconfigureSchedulerForWebsites()

	c.Run()
}

func (s *SchedulerServiceImpl) ScheduleWebsiteManual(websiteName string) {
	mongoInstance := helper.UgbMongoInstance

	col := mongoInstance.GetCollection(_const.UgbMongoDb, "pageRef")

	filters := bson.M{"tags": "reschedule", "websiteName": websiteName}

	update := bson.M{"$set": bson.M{"state": "DOWNLOAD", "status": "PENDING"}}

	res, err := col.UpdateMany(context.TODO(), filters, update)

	if err != nil {
		log.Print(err)
	}

	log.Printf("SchedulerServiceImpl sent %d items to deep scanning", res.ModifiedCount)
}

func (s *SchedulerServiceImpl) ReloadWebsites() {
	log.Print("reconfigure scheduler started")

	mongoInstance := helper.UgbMongoInstance

	col := mongoInstance.GetCollection("ug", "website")

	cursor, err := col.Find(context.TODO(), bson.M{})

	if err != nil {
		log.Print(err)
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		website := new(model.WebSite)

		err := cursor.Decode(website)

		if err != nil {
			log.Print(err)
			return
		}

		s.reconfigureSchedulerForWebsite(website)
		s.reconfigureEntryPoints(website)
	}

	log.Print("reconfigure scheduler done")
}

func (s *SchedulerServiceImpl) reconfigureSchedulerForWebsites() {
	log.Print("reconfigure scheduler started")

	mongoInstance := helper.UgbMongoInstance

	col := mongoInstance.GetCollection("ug", "website")

	cursor, err := col.Find(context.TODO(), bson.M{})

	if err != nil {
		log.Print(err)
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		website := new(model.WebSite)

		err := cursor.Decode(website)

		if err != nil {
			log.Print(err)
			return
		}

		s.reconfigureSchedulerForWebsite(website)
		s.reconfigureEntryPoints(website)
		s.runScheduler(website)
	}

	log.Print("reconfigure scheduler done")
}

func (s *SchedulerServiceImpl) reconfigureSchedulerForWebsite(website *model.WebSite) {
	s.tagMatchers[website.Name] = website.TagMatch
}

func (s *SchedulerServiceImpl) ConfigurePageRef(pageRef *model.PageRef) {
	s.applyTags(pageRef)

	if pageRef.Data.Status == "FINISHED" {
		s.ConfigureNextTask(pageRef)
	}
}

func (s *SchedulerServiceImpl) ConfigureNextTask(ref *model.PageRef) {
	if util.Contains(*ref.Data.Tags, "deep-scan") && ref.Data.State == "DOWNLOAD" {
		ref.Data.State = "DEEP_SCAN"
		ref.Data.Status = "PENDING"
	} else if util.Contains(*ref.Data.Tags, "allow-parse") && ref.Data.State == "DOWNLOAD" {
		ref.Data.State = "PARSE"
		ref.Data.Status = "PENDING"
	} else if util.Contains(*ref.Data.Tags, "allow-parse") && ref.Data.State == "DEEP_SCAN" {
		ref.Data.State = "PARSE"
		ref.Data.Status = "PENDING"
	} else if ref.Data.State == "PARSE" {
		//ref.Data.State = "PUBLISH"
		//ref.Data.Status = "PENDING"

		ref.Data.State = "PUBLISH"
		ref.Data.Status = "FINISHED"
	}
}

func (s *SchedulerServiceImpl) ConfigurePageUrl(pageRef *model.PageRef) {
	urlObj, err := url.Parse(pageRef.Data.Url)

	if err != nil {
		log.Panic(err)
	}

	rawQuery := urlObj.RawQuery
	if rawQuery == "" {
		return
	}

	query := urlObj.Query()
	urlObj.RawQuery = ""

	var queryKey = make(map[string]bool)
	for _, tag := range findSuffixTags(*pageRef.Data.Tags, "allow-query-") {
		param := tag[len("allow-query-"):]

		if query.Get(param) != "" {
			if queryKey[param] {
				continue
			}
			if len(urlObj.RawQuery) > 0 {
				urlObj.RawQuery += "&"
			}
			queryKey[param] = true

			urlObj.RawQuery += param + "=" + query.Get(param)
		}
	}

	rePrepareUrl(pageRef, urlObj.String())
}

func rePrepareUrl(ref *model.PageRef, newUrl string) {
	if ref.Data.Url == newUrl {
		return
	}

	//log.Printf("URL changed from %s to %s ; tags: %s", ref.Url, newUrl, strings.Join(*ref.Tags, ","))

	id, _ := uuid.FromString(helper.NamedUUID([]byte(newUrl)))
	ref.Id = id
	ref.Data.Url = newUrl
}

func (s *SchedulerServiceImpl) applyTags(pageRef *model.PageRef) {
	if s.tagMatchers[pageRef.Data.Source] == nil {
		return
	}

	if pageRef.Data.Tags == nil {
		pageRef.Data.Tags = &[]string{}
	}

	for _, tagMatcher := range s.tagMatchers[pageRef.Data.Source] {
		if s.checkPageRefMatchesPattern(tagMatcher, pageRef) {
			for _, tag := range tagMatcher.Tags {
				*pageRef.Data.Tags = append(*pageRef.Data.Tags, tag)
			}
		}
	}

	*pageRef.Data.Tags = unique(*pageRef.Data.Tags)
}

func unique(arr []string) []string {
	var set = make(map[string]bool)

	var result []string

	for _, item := range arr {
		if !set[item] {
			result = append(result, item)
		}

		set[item] = true
	}

	return result
}

func findSuffixTags(tags []string, prefix string) []string {
	var matchedTags []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			matchedTags = append(matchedTags, tag)
		}
	}

	return matchedTags
}

func (s *SchedulerServiceImpl) checkPageRefMatchesPattern(matcher model.TagMatcher, ref *model.PageRef) bool {
	if matcher.Pattern != "" {
		r := s.getRegexp(matcher.Pattern)

		if r != nil {
			res := r.Match([]byte(ref.Data.Url))

			if res {
				return true
			}
		}
	}

	for _, pattern := range matcher.Patterns {
		r := s.getRegexp(pattern)

		if r != nil {
			res := r.Match([]byte(ref.Data.Url))

			if res {
				return true
			}
		}
	}

	return false
}

func (s *SchedulerServiceImpl) getRegexp(pattern string) *regexp.Regexp {
	if s.regexCache[pattern] == nil {
		r, err := regexp.Compile(pattern)

		if err != nil {
			log.Print(err)
			return nil
		}

		s.regexCache[pattern] = r
	}

	return s.regexCache[pattern]
}

func (s *SchedulerServiceImpl) reconfigureEntryPoints(website *model.WebSite) {

	timeCalc := new(helper.TimeCalc)
	timeCalc.Init("pageRefApiList")

	var list []model.PageRef

	if len(website.EntryPoints) == 0 {
		return
	}

	for _, entryPoint := range website.EntryPoints {
		id, _ := uuid.FromString(helper.NamedUUID([]byte(entryPoint)))

		tags := append([]string{}, "entry-point", "deep-scan")
		pageRef := model.PageRef{
			Id: id,
			Data: model.PageRefData{
				Source: website.Name,
				Url:    entryPoint,
				State:  "TODO",
				Status: "PENDING",
				Tags:   &tags,
			},
		}

		s.ConfigurePageRef(&pageRef)

		if !s.service.PageRefExists(id) {
			list = append(list, pageRef)
		}
	}

	if len(list) > 0 {
		s.service.BulkWrite(list)
	}
}

func (s *SchedulerServiceImpl) runScheduler(website *model.WebSite) {
	for _, schedule := range website.Schedule {
		s.runScheduleOnWebsite(website, schedule)
	}
}

func (s *SchedulerServiceImpl) runScheduleOnWebsite(website *model.WebSite, schedule model.WebsiteSchedule) {
	schedule.Match.WebsiteName = website.Name

	pageChan, updateChan := s.service.UpdateStatesBulk2(context.TODO(), schedule.Match, schedule.ToState, schedule.ToStatus)
	defer close(updateChan)

	for pageRef := range pageChan {
		updateChan <- pageRef
	}
}
