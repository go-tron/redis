package redis

import (
	"context"
	"encoding/json"
	"github.com/go-tron/local-time"
	"strconv"
	"time"
)

type Duration int

const (
	DurationHour Duration = 1
	DurationDay  Duration = 2
)

type Statistics struct {
	Redis    *Redis
	Key      string
	Duration Duration
}

func (s *Statistics) Save(ctx context.Context, id string) error {
	var expireIn time.Duration
	var dateformat string
	if s.Duration == DurationHour {
		dateformat = "1504"
		expireIn = time.Hour
	} else if s.Duration == DurationDay {
		dateformat = "0215"
		expireIn = time.Hour * 24
	}
	t := localTime.Now().Format(dateformat)
	_, err := s.Redis.FrequencyLimit(ctx, s.Key+":"+id+":"+t, 0, expireIn)
	return err
}

func (s *Statistics) Get(ctx context.Context, id string) (int, error) {
	var fields []interface{}
	if s.Duration == DurationHour {
		for i := 0; i < 60; i++ {
			fields = append(fields, localTime.Now().Add(-time.Minute*time.Duration(i)).Format("1504"))
		}
	} else if s.Duration == DurationDay {
		for i := 0; i < 23; i++ {
			fields = append(fields, localTime.Now().Add(-time.Hour*time.Duration(i)).Format("0215"))
		}
	}

	result, err := s.Redis.BatchGet(ctx, s.Key+":"+id, fields...)
	if err != nil {
		return 0, err
	}

	var resultData []string
	if err := json.Unmarshal([]byte(result.(string)), &resultData); err != nil {
		return 0, err
	}

	var total = 0
	for _, val := range resultData {
		v, err := strconv.Atoi(val)
		if err != nil {
			continue
		}
		total += v
	}
	return total, nil
}
