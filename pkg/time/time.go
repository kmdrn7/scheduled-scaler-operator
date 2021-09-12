package time

import (
	"fmt"
	"time"
)

var (
	timeLayout    = "2006-01-02T15:04:05Z"
	timeZone, err = time.LoadLocation("Asia/Jakarta")
)

func init() {
	if err != nil {
		fmt.Println("Error configuring timezone")
		panic(err)
	}
}

func Now() time.Time {
	return time.Now().UTC().In(timeZone)
}

func Parse(timeString string) (time.Time, error) {
	parsedTime, err := time.ParseInLocation(timeLayout, timeString, timeZone)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}
