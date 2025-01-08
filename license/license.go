package license

import (
	"time"
)

// 过期时间当地时间2021-01-01 00:00:00
var expireTime time.Time = time.Date(2100, 3, 31, 0, 0, 0, 0, time.Local)

// const (
// 	ntpServer string = "ntp1.aliyun.com"
// )

// var networkTime time.Time

// GetExpireTime ...
func GetExpireTime() *time.Time {
	return &expireTime
}

// func init() {
// 	// 每小时同步一次
// 	ticker := time.Tick(time.Hour)
// 	for {
// 		<-ticker
// 		var err error
// 		networkTime, err = ntp.Time(ntpServer)
// 		if err != nil {
// 			logger.CAP.Error("同步网络时间失败：" + err.Error())
// 		}
// 	}

// }

// TimeSource 时间来源
type TimeSource int

const (
	// TimeSourceNetwork 网络时间
	TimeSourceNetwork = iota
	// TimeSourceLocalhost 本机时间
	TimeSourceLocalhost
)

// Time ...
type Time struct {
	time.Time
	Source TimeSource
}

// Now ...
func Now() (*Time, error) {
	return &Time{
		Source: TimeSourceLocalhost,
		Time:   time.Now(),
	}, nil
}
