package redis

import (
	"context"
	"testing"
	"time"
)

var s = Statistics{
	Redis: New(&Config{
		Addr:     "127.0.0.1:6379",
		Password: "GBkrIO9bkOcWrdsC",
	}),
	Key:      "test-01",
	Duration: DurationHour,
}

func TestStatisticsSave(t *testing.T) {
	for i := 0; i < 100; i++ {
		err := s.Save(context.Background(), "1")
		t.Log(i, err)
		time.Sleep(time.Second)
	}
}
func TestStatisticsGet(t *testing.T) {
	val, err := s.Get(context.Background(), "1")
	t.Log(val, err)
}
