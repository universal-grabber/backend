package helper

import (
	"backend/api/model"
	"time"
)

func BufferChan(input chan *model.PageRef) chan []*model.PageRef {
	pageRefUpdateChanBuffered := make(chan []*model.PageRef)

	go func() {
		var buffer []*model.PageRef
		var isClosed = false

		go func() {
			for !isClosed {
				time.Sleep(100 * time.Millisecond)
				if len(buffer) == 0 {
					continue
				}
				var temp = buffer
				buffer = []*model.PageRef{}
				pageRefUpdateChanBuffered <- temp
			}
			close(pageRefUpdateChanBuffered)
		}()

		for item := range input {
			buffer = append(buffer, item)
		}

		isClosed = true
	}()

	return pageRefUpdateChanBuffered
}
