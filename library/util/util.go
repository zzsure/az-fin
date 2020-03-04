package util

import (
	"az-fin/library/util/net"
	"encoding/binary"
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

func GetMillTimeByDate(date string) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	return t.Unix() * 1000, err
}

func GetFormatTime(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc).Format("2006-01-02 15:04:05")
}

func GetDateByTime(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(loc).Format("20060102")
}

func GetUnixTime() int64 {
	return time.Now().Unix()
}

func GetMillUnixTime() int64 {
	return time.Now().UnixNano() / 1e6
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func GetMorningUnixTime(t time.Time) int64 {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Unix()
}
