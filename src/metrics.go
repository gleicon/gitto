package main

import (
	"fmt"
	"github.com/fiorix/go-redis/redis"
	"strings"
	"time"
)

type Metrics struct {
	rc *redis.Client
}

func NewMetrics(rc *redis.Client) *Metrics {
	mm := Metrics{}
	mm.rc = rc
	return &mm
}

func (mm *Metrics) CounterIncr(application string, name string, value int) (int, error) {
	key := fmt.Sprintf("gitto:%s:counters", strings.ToLower(application))
	err := mm.StoreTSPoint(application, name, value)
	if err != nil {
		return 0, err
	}
	return mm.rc.HIncrBy(key, name, value)
}

func (mm *Metrics) GetCounters(application string) (map[string]string, error) {
	key := fmt.Sprintf("gitto:%s:counters", strings.ToLower(application))
	return mm.rc.HGetAll(key)
}

func (mm *Metrics) AppendRequest(application string, request string) (int, error) {
	key := fmt.Sprintf("gitto:%s:requests", strings.ToLower(application))
	return mm.rc.LPush(key, request)
}

func (mm *Metrics) GetRequests(application string, limit int) ([]string, error) {
	key := fmt.Sprintf("gitto:%s:requests", strings.ToLower(application))
	return mm.rc.LRange(key, 0, limit)
}

func (mm *Metrics) StoreTSPoint(application string, tsname string, value int) error {
	tt := getTimeNow()
	key := fmt.Sprintf("gitto:%s:ts:%s", strings.ToLower(application), strings.ToLower(tsname))
	_, err := mm.rc.HIncrBy(key, tt, value)
	return err
}

func (mm *Metrics) FetchTS(application string, tsname string) (map[string]string, error) {
	key := fmt.Sprintf("gitto:%s:ts:%s", strings.ToLower(application), strings.ToLower(tsname))
	return mm.rc.HGetAll(key)
}

/*
	getTimeNow() returns a string representing YYYY:MM:DD:HH:MM (minute granularity)
*/

func getTimeNow() string {
	n := time.Now()
	tt := fmt.Sprintf("%d%02d:%02dT%02d:%02d:%02d\n",
		n.Year(), n.Month(), n.Day(),
		n.Hour(), n.Minute(), n.Second())
	return tt
}
