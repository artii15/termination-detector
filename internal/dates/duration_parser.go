package dates

import (
	"fmt"
	"time"
)

func MustParseDuration(durationString string) time.Duration {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		panic(fmt.Sprintf("failed to parse duration: %s\n%s", err.Error(), durationString))
	}
	return duration
}
