package service

import log "github.com/sirupsen/logrus"

type PageRefService struct {
}

func (service *PageRefService) Test1() {
	log.Print("test1 called")
}
