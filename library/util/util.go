package util

import (
	"az-fin/library/util/net"
	"encoding/json"
	"fmt"
	"strings"
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
