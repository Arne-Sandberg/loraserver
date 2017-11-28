package downlink

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/loraserver/internal/common"
	"github.com/brocaar/loraserver/internal/storage"
)

// ClassCSchedulerLoop starts an infinit loop calling the Class-C scheduler
// each Class-C schedule interval.
func ClassCSchedulerLoop() {
	for {
		log.Debug("running class-c scheduler batch")
		if err := ClassCScheduleBatch(common.ClassCScheduleBatchSize); err != nil {
			log.WithError(err).Error("class-c scheduler error")
		}
		time.Sleep(common.ClassCScheduleInterval)
	}
}

func ClassCScheduleBatch(size int) error {
	return storage.Transaction(common.DB, func(tx *sqlx.Tx) error {
		devices, err := storage.GetDevicesWithClassCDeviceQueueItems(tx, size)
		if err != nil {
			return errors.Wrap(err, "get deveuis with class-c device-queue items error")
		}

		for _, d := range devices {
			ds, err := storage.GetDeviceSession(common.RedisPool, d.DevEUI)
			if err != nil {
				log.WithError(err).WithField("dev_eui", d.DevEUI).Error("get device-session error")
				continue
			}

			err = Flow.RunScheduleNextDeviceQueueItem(ds)
			if err != nil {
				log.WithError(err).WithField("dev_eui", d.DevEUI).Error("schedule next device-queue item error")
			}
		}

		return nil
	})
}
