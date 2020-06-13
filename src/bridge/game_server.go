package bridge

import (
	"common"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HttpGetServer(gameId int32, method string, para string) string {
	return httpDoServer("GET", gameId, method, para)
}

func HttpPostServer(gameId int32, method string, para string) string {
	return httpDoServer("POST", gameId, method, para)
}

func HttpGetServer2(gameId int32, method string, mPara map[string]string) string {
	return httpDoServer2("GET", gameId, method, mPara)
}

func HttpPostServer2(gameId int32, method string, mPara map[string]string) string {
	return httpDoServer2("POST", gameId, method, mPara)
}

func httpDoServer(do string, gameId int32, method string, para string) string {
	var err error
	var resp *http.Response
	if do == "GET" {
		resp, err = http.Get(fmt.Sprintf("http://%s:9131/rest/stat/%s?%s", common.GetAllServerIP()[gameId], method, para))
	} else if do == "POST" {
		resp, err = http.Post(fmt.Sprintf("http://%s:9131/rest/stat/%s", common.GetAllServerIP()[gameId], method),
			"application/x-www-form-urlencoded", strings.NewReader(para))
	} else {
		return ""
	}
	if err != nil {
		logs.Error("httpDoServer Do error:", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("httpDoServer ReadAll error:", err)
		return ""
	}
	return string(body)
}

func httpDoServer2(do string, gameId int32, method string, mPara map[string]string) string {
	var para string
	for k, v := range mPara {
		para += k + "=" + url.QueryEscape(v) + "&"
	}
	if para != "" {
		para = para[:len(para)-1]
	}
	return httpDoServer(do, gameId, method, para)
}
