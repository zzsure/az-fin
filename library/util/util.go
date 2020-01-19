package util

import (
	"az-fin/library/util/net"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func GenServerUUID() string {
	ip, mac := net.NewLAN().NetInfo()
	return fmt.Sprintf("%s-%s", ip, mac)
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func GetURL(baseURL, uri string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), strings.TrimLeft(uri, "/"))
}

// random sleep
func RandomSleep(t int) {
	rs := rand.Intn(t)
	time.Sleep(time.Duration(rs) * time.Second)
}

func GetTimeByMillUnixTime(t int64) time.Time {
	t = t / 1000
	return time.Unix(t, 0)
}

func GetFormatTime(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc).Format("2006-01-02 15:04:05")
}

func GetMillUnixTime() int64 {
	return time.Now().UnixNano() / 1e6
}
