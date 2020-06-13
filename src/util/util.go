package util

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

func Dec2letter(n uint32) string {
	var str []byte
	for {
		str = append(str, byte(n%26)+byte('a'))
		if n < 26 {
			break
		}
		n = n/26 - 1
	}
	ret := make([]byte, len(str))
	for i, v := range str {
		ret[len(str)-i-1] = v
	}
	return string(ret)
}

func Letter2dec(str string) uint32 {
	var n uint32
	for i, v := range str {
		n += uint32(v-'a'+1) * uint32(math.Pow(26, float64(len(str)-i-1)))
	}
	return n - 1
}

func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func Date2ts(date string) int64 {
	t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	return t.Unix()
}

func Time2ts(dt string) int64 {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
	return t.Unix()
}

func Ts2date(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02")
}

func Ts2time(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func GetDate() string {
	return Ts2date(time.Now().Unix())
}

func GetTime() string {
	return Ts2time(time.Now().Unix())
}

func RemoveBlank(s string) string {
	var c byte
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' || s[i] == '"' {
			if c == 0 {
				c = s[i]
			} else if s[i] == c {
				c = 0
			}
		} else if c == 0 && (s[i] == ' ' || s[i] == '	') {
			s = s[:i] + s[i+1:]
			i--
		}
	}
	return s
}

func RemoveAllBlank(s string) string {
	s = strings.Replace(s, "	", "", -1)
	s = strings.Replace(s, " ", "", -1)
	return s
}

func RemoveLeftBlank(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '	' {
			s = s[i:]
			break
		}
	}
	return s
}

func RemoveRightBlank(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' && s[i] != '	' {
			s = s[:i+1]
			break
		}
	}
	return s
}

func RemoveSideBlank(s string) string {
	return RemoveRightBlank(RemoveLeftBlank(s))
}

func InSlice(slice interface{}, elem interface{}) bool {
	if reflect.TypeOf(slice).Kind() != reflect.Slice || reflect.TypeOf(slice).Elem() != reflect.TypeOf(elem) {
		return false
	}
	if t := reflect.TypeOf(elem).Kind(); !(t >= reflect.Int && t <= reflect.Int64 || t >= reflect.Uint && t <= reflect.Uint64 ||
		t == reflect.Float32 || t == reflect.Float64 || t == reflect.String) {
		return false
	}
	valSlice := reflect.ValueOf(slice)
	valElem := reflect.ValueOf(elem)
	for i := 0; i < valSlice.Len(); i++ {
		if valSlice.Index(i).Interface() == valElem.Interface() {
			return true
		}
	}
	return false
}

func InSlice2(slice interface{}, equal func(i int) bool) bool {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return false
	}
	valSlice := reflect.ValueOf(slice)
	for i := 0; i < valSlice.Len(); i++ {
		if equal(i) {
			return true
		}
	}
	return false
}

func RemoveSliceElem(slice interface{}, elem interface{}, all bool) interface{} {
	if reflect.TypeOf(slice).Kind() != reflect.Slice || reflect.TypeOf(slice).Elem() != reflect.TypeOf(elem) {
		return slice
	}
	if t := reflect.TypeOf(elem).Kind(); !(t >= reflect.Int && t <= reflect.Int64 || t >= reflect.Uint && t <= reflect.Uint64 ||
		t == reflect.Float32 || t == reflect.Float64 || t == reflect.String) {
		return slice
	}
	valSlice := reflect.ValueOf(slice)
	valElem := reflect.ValueOf(elem)
	for i := 0; i < valSlice.Len(); i++ {
		if valSlice.Index(i).Interface() == valElem.Interface() {
			valSlice = reflect.AppendSlice(valSlice.Slice(0, i), valSlice.Slice(i+1, valSlice.Len()))
			if all {
				i--
			} else {
				break
			}
		}
	}
	return valSlice.Interface()
}

func RemoveSliceElem2(slice interface{}, equal func(i int) bool, all bool) interface{} {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return slice
	}
	valSlice := reflect.ValueOf(slice)
	for i := 0; i < valSlice.Len(); i++ {
		if equal(i) {
			valSlice = reflect.AppendSlice(valSlice.Slice(0, i), valSlice.Slice(i+1, valSlice.Len()))
			if all {
				i--
			} else {
				break
			}
		}
	}
	return valSlice.Interface()
}

func UniqueSlice(slice interface{}, bSort bool) interface{} {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return slice
	}
	if t := reflect.TypeOf(slice).Elem().Kind(); !(t >= reflect.Int && t <= reflect.Int64 || t >= reflect.Uint && t <= reflect.Uint64 ||
		t == reflect.Float32 || t == reflect.Float64 || t == reflect.String) {
		return slice
	}
	valSlice := reflect.ValueOf(slice)
	if valSlice.Len() < 2 {
		return slice
	}
	for i := 0; i < valSlice.Len(); i++ {
		bDel := false
		if bSort {
			if i > 0 && valSlice.Index(i-1).Interface() == valSlice.Index(i).Interface() {
				bDel = true
			}
		} else {
			for j := 0; j < i; j++ {
				if valSlice.Index(j).Interface() == valSlice.Index(i).Interface() {
					bDel = true
					break
				}
			}
		}
		if bDel {
			valSlice = reflect.AppendSlice(valSlice.Slice(0, i), valSlice.Slice(i+1, valSlice.Len()))
			i--
		}
	}
	return valSlice.Interface()
}

func UniqueSlice2(slice interface{}, equal func(i, j int) bool, bSort bool) interface{} {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return slice
	}
	valSlice := reflect.ValueOf(slice)
	if valSlice.Len() < 2 {
		return slice
	}
	for i := 0; i < valSlice.Len(); i++ {
		bDel := false
		if bSort {
			if i > 0 && equal(i-1, i) {
				bDel = true
			}
		} else {
			for j := 0; j < i; j++ {
				if equal(j, i) {
					bDel = true
					break
				}
			}
		}
		if bDel {
			valSlice = reflect.AppendSlice(valSlice.Slice(0, i), valSlice.Slice(i+1, valSlice.Len()))
			i--
		}
	}
	return valSlice.Interface()
}

func Slice2Map(slice interface{}) map[interface{}]int {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return nil
	}
	m := make(map[interface{}]int)
	valSlice := reflect.ValueOf(slice)
	for i := 0; i < valSlice.Len(); i++ {
		m[valSlice.Index(i).Interface()] = i
	}
	return m
}

func ToJson(i interface{}) string {
	data, _ := json.Marshal(i)
	return string(data)
}

func FromJson(s string, i interface{}) error {
	return json.Unmarshal([]byte(s), i)
}

func GobEncodeFile(filename string, dt interface{}) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(dt); err != nil {
		logs.Error("GobEncodeFile, Encode error:", err)
		return
	}
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		logs.Error("GobEncodeFile, WriteFile error:", err)
		return
	}
}

func GobDecodeFile(filename string, dt interface{}) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		logs.Error("GobDecodeFile, ReadFile error:", err)
		return
	}
	err = gob.NewDecoder(bytes.NewBuffer(buf)).Decode(dt)
	if err != nil {
		logs.Error("GobDecodeFile, Decode error:", err)
		return
	}
}

func SaveToFile(r io.Reader, filename string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logs.Error("[SaveToFile] OpenFile error:", err)
		return err
	}
	defer f.Close()
	for {
		buf := make([]byte, 4096)
		n, err := r.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				logs.Error("[SaveToFile] Write error:", err)
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logs.Error("[SaveToFile] Read error:", err)
				return err
			}
		}
	}
	return nil
}

func HttpListen(addr string) {
	go func() {
		defer func() {
			msg := recover()
			if msg != "" {
				logs.Error("http server abnormal with", msg)
			} else {
				logs.Notice("http server stopped!")
			}
		}()
		http.ListenAndServe(addr, nil)
	}()
}

func HttpsListen(addr string, cert, key string) {
	go func() {
		defer func() {
			msg := recover()
			if msg != "" {
				logs.Error("https server abnormal with", msg)
			} else {
				logs.Notice("https server stopped!")
			}
		}()
		http.ListenAndServeTLS(addr, cert, key, nil)
	}()
}

var g_uuid uint32

func GetUUID() uint32 {
	return atomic.AddUint32(&g_uuid, 1)
}
