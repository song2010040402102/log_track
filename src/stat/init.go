package stat

import (
	"config"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
	"text/template"
	"util"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		logs.Info("IndexHandler")
		t, err := template.ParseFiles("../template/index.html")
		if err == nil {
			t.Execute(w, nil)
		} else {
			fmt.Fprintln(w, err.Error())
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func Stat2JsonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gameId, _ := strconv.ParseInt(r.PostFormValue("game_id"), 10, 32)
		item := r.PostFormValue("item")
		start := r.PostFormValue("start")
		end := r.PostFormValue("end")
		channel := r.PostFormValue("channel")
		device := r.PostFormValue("device")
		logs.Info("Stat2JsonHandler, gameId:", gameId, "item:", item, "start:", start, "end:", end, "channel:", channel, "device:", device)
		if itemRes := GetItemResult(int32(gameId), item, start, end, channel, device); itemRes != nil {
			fmt.Fprintln(w, util.ToJson(itemRes))
		} else {
			fmt.Fprintln(w, "no result!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func ResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gameId, _ := strconv.ParseInt(r.PostFormValue("game_id"), 10, 32)
		item := r.PostFormValue("item")
		items := r.PostFormValue("items")
		start := r.PostFormValue("start")
		end := r.PostFormValue("end")
		channel := r.PostFormValue("channel")
		device := r.PostFormValue("device")
		output := r.PostFormValue("output")
		logs.Info("ResultHandler, gameId:", gameId, "item:", item, "items:", items, "start:", start, "end:", end, "channel:", channel, "device:", device, "output:", output)

		var itemRes *ItemResult
		if items == "" {
			itemRes = GetItemResult(int32(gameId), item, start, end, channel, device)
			AddIndexToItemRes(itemRes)
		} else {
			itemRes = GetItemResult2(int32(gameId), items, start, end, channel, device)
		}
		if itemRes != nil {
			if itemRes.Ret != 0 {
				fmt.Fprintln(w, "[ERROR]"+itemRes.Error())
			} else if output == "web" {
				t, err := template.ParseFiles("../template/result.html", "../template/table.html")
				if err == nil {
					t.Execute(w, itemRes)
				} else {
					fmt.Fprintln(w, "[ERROR]"+err.Error())
				}
			} else if output == "xlsx" {
				filename, err := WriteXLSX(itemRes)
				if err == nil {
					fmt.Fprintln(w, "download/"+filename)
				} else {
					fmt.Fprintln(w, "[ERROR]"+err.Error())
				}
			} else {
				fmt.Fprintln(w, "[ERROR]please select one lookup method at least!")
			}
		} else {
			fmt.Fprintln(w, "[ERROR]no result!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "[ERROR]not support!")
	}
}

func ShowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gameId, _ := strconv.ParseInt(r.PostFormValue("game_id"), 10, 32)
		flow := r.PostFormValue("flow")
		start := r.PostFormValue("start_time")
		end := r.PostFormValue("end_time")
		role := r.PostFormValue("role_name")
		filter := r.PostFormValue("filter")
		output := r.PostFormValue("output")
		logs.Info("ShowHandler, gameId:", gameId, "flow:", flow, "start:", start, "end:", end, "role:", role, "filter:", filter, "output:", output)
		if flowData := ShowFlow(int32(gameId), flow, start, end, role, filter); flowData != nil {
			if flowData.Ret != 0 {
				fmt.Fprintln(w, "[ERROR]"+flowData.Error())
			} else if output == "web" {
				t, err := template.ParseFiles("../template/result.html", "../template/table.html")
				if err == nil {
					t.Execute(w, flowData)
				} else {
					fmt.Fprintln(w, "[ERROR]"+err.Error())
				}
			} else if output == "xlsx" {
				filename, err := WriteXLSX(flowData)
				if err == nil {
					fmt.Fprintln(w, "download/"+filename)
				} else {
					fmt.Fprintln(w, "[ERROR]"+err.Error())
				}
			} else {
				fmt.Fprintln(w, "[ERROR]please select one lookup method at least!")
			}
		} else {
			fmt.Fprintln(w, "[ERROR]no result!")
		}
	} else {
		fmt.Fprintln(w, "[ERROR]"+r.Method, "not support!")
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	const maxSize = 100 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		fmt.Fprintln(w, "[ERROR]file too big!")
		return
	}
	tp, _ := strconv.ParseInt(r.PostFormValue("third_party"), 10, 32)
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, "[ERROR]"+err.Error())
		return
	}
	defer file.Close()
	filename := fmt.Sprintf("./upload/temp_%d.xlsx", util.GetUUID())
	err = util.SaveToFile(file, filename)
	if err != nil {
		fmt.Fprintln(w, "[ERROR]"+err.Error())
		return
	}
	logs.Info("UploadHandler, tp:", tp, "filename:", filename)
	err = data.HandleThirdPartyData(int32(tp), filename)
	if err != nil {
		fmt.Fprintln(w, "[ERROR]"+err.Error())
		return
	}
	fmt.Fprintln(w, "success!")
}

func RelationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gameId, _ := strconv.ParseInt(r.PostFormValue("game_id"), 10, 32)
		item := r.PostFormValue("item")
		start := r.PostFormValue("start_time")
		end := r.PostFormValue("end_time")
		output := r.PostFormValue("output")
		logs.Info("RelationHandler, gameId:", gameId, "item:", item, "start:", start, "end:", end, "output:", output)
		if relationData := GetRelation(int32(gameId), item, start, end); relationData != nil {
			if relationData.Ret != 0 {
				fmt.Fprintln(w, "[ERROR]"+relationData.Error())
			} else if output == "web" {
				t, err := template.ParseFiles("../template/result.html", "../template/table.html")
				if err == nil {
					t.Execute(w, relationData)
				} else {
					fmt.Fprintln(w, "[ERROR]"+err.Error())
				}
			} else if output == "xlsx" {
				filename, err := WriteXLSX(relationData)
				if err == nil {
					fmt.Fprintln(w, "download/"+filename)
				} else {
					fmt.Fprintln(w, "[ERROR]"+err.Error())
				}
			} else {
				fmt.Fprintln(w, "[ERROR]please select one lookup method at least!")
			}
		} else {
			fmt.Fprintln(w, "[ERROR]no result!")
		}
	} else {
		fmt.Fprintln(w, "[ERROR]"+r.Method, "not support!")
	}
}

func HandleDownload(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
}

func StartServer() {
	data.Init()
	logs.Notice("http server start...")
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/stat2json", Stat2JsonHandler)
	http.HandleFunc("/result", ResultHandler)
	http.HandleFunc("/show", ShowHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/relation", RelationHandler)
	http.Handle("/download/", HandleDownload(http.StripPrefix("/download", http.FileServer(http.Dir("download")))))
	http.Handle("/html/", http.StripPrefix("/html", http.FileServer(http.Dir("../template"))))
	if config.Get().Listen.Stat != "" {
		util.HttpListen(config.Get().Listen.Stat)
	}
}

func CloseServer() {
	data.Flush()
}
