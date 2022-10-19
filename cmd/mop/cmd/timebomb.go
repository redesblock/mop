package cmd

import (
	ver "github.com/redesblock/mop"
	"strconv"
	"time"

	"github.com/redesblock/mop/core/log"
)

const (
	limitDays   = 90
	warningDays = 0.9 * limitDays // show warning once 90% of the time bomb time has passed
	sleepFor    = 30 * time.Minute
)

var (
	commitTime, _   = strconv.ParseInt(ver.CommitTime(), 10, 64)
	versionReleased = time.Unix(commitTime, 0)
)

func startTimeBomb(logger log.Logger) {
	for {
		outdated := time.Now().AddDate(0, 0, -limitDays)

		if versionReleased.Before(outdated) {
			logger.Warning("your node is outdated, please check for the latest version")
		} else {
			almostOutdated := time.Now().AddDate(0, 0, -warningDays)

			if versionReleased.Before(almostOutdated) {
				logger.Warning("your node is almost outdated, please check for the latest version")
			}
		}

		<-time.After(sleepFor)
	}
}

func endSupportDate() string {
	return versionReleased.AddDate(0, 0, limitDays).Format("2 January 2006")
}
