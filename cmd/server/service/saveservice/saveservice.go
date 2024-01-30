package saveservice

import (
	"context"
	"log"
	"time"

	"github.com/AndreyVLZ/metrics/cmd/server/producer"
	"github.com/AndreyVLZ/metrics/internal/storage"
)

type saveSevice struct {
	store    storage.Storage
	producer *producer.Producer
	storeInt int
	exit     chan struct{}
}

func New(store storage.Storage, storeInt int, producer *producer.Producer) *saveSevice {
	return &saveSevice{
		store:    store,
		producer: producer,
		storeInt: storeInt,
		exit:     make(chan struct{}),
	}
}

func (ss *saveSevice) Name() string { return "saveService" }

func (ss *saveSevice) Start() error {
	err := ss.producer.Open()
	if err != nil {
		return err
	}

	go ss.start()

	return nil
}

func (ss *saveSevice) start() {
	for {
		select {
		case <-time.After(time.Duration(ss.storeInt) * time.Second):
			if err := ss.saved(); err != nil {
				log.Printf("err save metrics %v\n", err)
			}
		case <-ss.exit:
			return
		}
	}
}

func (ss *saveSevice) saved() error {
	arr := ss.store.List(context.Background())
	if len(arr) == 0 {
		return nil
	}

	err := ss.producer.Trunc()
	if err != nil {
		return err
	}

	for _, m := range arr {
		err := ss.producer.WriteMetric(&m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ss *saveSevice) Stop() error {
	ss.exit <- struct{}{}

	err := ss.saved()
	if err != nil {
		return err
	}

	return ss.producer.Close()
}
