package main

import (
	"fmt"
	"strings"
	"time"
)

func main() {

	//repeated := true
	timeLayout := "2006-01-02T15:04:05Z"
	tz, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("Error set timezone")
	}

	now := time.Now().UTC().In(tz)

	schedule := "2021-09-09T11:00:00Z,2021-09-09T13:00:00Z"
	scheduleSplit := strings.Split(schedule, ",")
	scheduleStart := scheduleSplit[0]
	scheduleEnd := scheduleSplit[1]

	timeStart, err := time.ParseInLocation(timeLayout, scheduleStart, tz)
	if err != nil {
		fmt.Println("Error parsing time")
	}

	timeEnd, err := time.ParseInLocation(timeLayout, scheduleEnd, tz)
	if err != nil {
		fmt.Println("Error parsing time")
	}

	if now.After(timeStart) {
		fmt.Println("Waktu sudah mulai")
	}

	if now.After(timeEnd) {
		fmt.Println("Waktu sudah mulai")
	}

	if now.Before(timeStart) {
		// reconcile after specific time
		reconcileAfter := timeStart.Sub(now).Seconds()
		fmt.Println("Time to reconcile is", reconcileAfter)
	} else if now.After(timeStart) && now.Before(timeEnd) {
		// scale deployment
		// update status StoredReplicaCount to match with Deployment's replica count
		// <code here>
		// update status Phase to RUNNING
		// <code here>
		// reconcile again after specific time
		reconcileAfter := timeEnd.Sub(now).Seconds()
		fmt.Println("Time to reconcile is", reconcileAfter)
	} else if now.After(timeEnd) {
		// scale back deployment's replicaset to it's real value
		//<code here>
		// update status Phase to DONE
	}

	//fmt.Println(now)
	//fmt.Println(timeStart)
	//future := "2021-09-07T18:20:00Z"
	//
	//timeFuture, err := time.Parse(timeLayout, future)
	//
	//fmt.Println(now)
	//fmt.Println(timeFuture)
	//
	//if err != nil {
	//	fmt.Println("Error")
	//}
	//
	//fmt.Println(timeFuture.Sub(now))

	//fmt.Println("jarak : ", time.Now().After(time.))
	//fmt.Println(time)
}
