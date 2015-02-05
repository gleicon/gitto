package main

import (
	"fmt"
	"github.com/fiorix/go-redis/redis"
	"strings"
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
	key := fmt.Sprintf("gitto:counters:%s", strings.ToLower(application))
	return mm.rc.HIncrBy(key, name, value)
}

func (mm *Metrics) GetCounters(application string) (map[string]string, error) {
	key := fmt.Sprintf("gitto:counters:%s", strings.ToLower(application))
	return mm.rc.HGetAll(key)
}

func (mm *Metrics) AppendRequest(application string, request string) (int, error) {
	key := fmt.Sprintf("gitto:requests:%s", strings.ToLower(application))
	return mm.rc.LPush(key, request)
}

func (mm *Metrics) GetRequests(application string, limit int) ([]string, error) {
	key := fmt.Sprintf("gitto:requests:%s", strings.ToLower(application))
	return mm.rc.LRange(key, 0, limit)
}
