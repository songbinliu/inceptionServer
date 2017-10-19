package util

import (
	"time"
	"github.com/golang/glog"
)

func TimeTrack(start time.Time, name string) time.Duration{
	elapsed := time.Since(start)
	glog.V(2).Infof("%s took %s", name, elapsed)
	return elapsed
}