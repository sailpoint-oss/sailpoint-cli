package cmd

import (
	"testing"
	"time"
)

func Test_updateLastSeenTime(t *testing.T) {
	now := time.Now()
	if nextFromTime().UnixMilli() >= now.UnixMilli() {
		t.Errorf("unexepected first time returned")
	}
	for i := 0; i < 10; i++ {
		updateLastSeenTime(now)
		if i < 10 {
			now = time.Now()
		}
	}
	if nextFromTime().UnixMilli() != now.Add(1*time.Millisecond).UnixMilli() {
		t.Errorf("unexepected next time returned")
	}
}

func Test_updateLastSeenTimeOutOfSequence(t *testing.T) {
	first := time.Now()
	time.Sleep(1 * time.Millisecond)
	second := time.Now()
	time.Sleep(1 * time.Millisecond)
	now := time.Now()
	for i := 0; i < 10; i++ {
		updateLastSeenTime(now)
		if i < 10 {
			now = time.Now()
		}
	}
	updateLastSeenTime(first)
	updateLastSeenTime(second)
	if nextFromTime().UnixMilli() != now.Add(1*time.Millisecond).UnixMilli() {
		t.Errorf("unexepected next time returned")
	}
}
