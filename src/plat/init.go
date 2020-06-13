package plat

import (
	"config"
	"db"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
	"text/template"
	"util"
	"version"
	"xojoc.pw/useragent"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		logs.Info("IndexHandler")
		t, err := template.ParseFiles("../template/plat.html")
		if err == nil {
			t.Execute(w, nil)
		} else {
			fmt.Fprintln(w, err.Error())
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func LandPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		r.ParseForm()
		clientIP := r.Header.Get("X-Real-IP")
		channel := r.PostFormValue("channel")
		device := ""
		if ua := useragent.Parse(r.Header.Get("User-Agent")); ua != nil {
			device = fmt.Sprintf("%s_%d.%d", ua.OS, ua.OSVersion.Major, ua.OSVersion.Minor)
		}
		event := r.PostFormValue("event")
		logs.Info("LandPageHandler, clientIP:", clientIP, "channel:", channel, "device:", device, "event:", event)
		AddLandPageData(clientIP, channel, device, event)
		fmt.Fprintln(w, "success")
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func LandPageResHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		channel := r.Form.Get("channel")
		device := r.Form.Get("device")
		start, _ := strconv.ParseInt(r.Form.Get("start"), 10, 64)
		end, _ := strconv.ParseInt(r.Form.Get("end"), 10, 64)
		logs.Info("LandPageResHandler, channel:", channel, "device:", device, "start:", start, "end:", end)
		if lpRes := GetLandPageRes(channel, device, start, end); lpRes != nil {
			fmt.Fprintln(w, util.ToJson(lpRes))
		} else {
			fmt.Fprintln(w, "no result!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func TransAppHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" || r.Method == "POST" {
		r.ParseForm()
		cid := r.Form.Get("cid")
		os, deviceId, ts, callback_url := ParsePlatInfo(cid, r)
		logs.Info("TransAppHandler, cid:", cid, "os:", os, "deviceId:", deviceId, "ts:", ts, "callback_url:", callback_url)
		AddTransAppData(cid, os, deviceId, ts, callback_url)
		fmt.Fprintln(w, "success")
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func TransEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		channel := r.Form.Get("channel")
		os, _ := strconv.ParseInt(r.Form.Get("os"), 10, 32)
		deviceId := ""
		if os == 0 {
			deviceId = r.Form.Get("imei") + "##" + r.Form.Get("androidid")
		} else {
			deviceId = r.Form.Get("idfa")
		}
		event, _ := strconv.ParseInt(r.Form.Get("event"), 10, 32)
		logs.Info("TransEventHandler, channel:", channel, "os:", os, "deviceId:", deviceId, "event:", event)
		UpdateTransEvent(channel, int8(os), deviceId, int8(event))
		fmt.Fprintln(w, "success")
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func TransResHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		channel := r.Form.Get("channel")
		device := r.Form.Get("device")
		start, _ := strconv.ParseInt(r.Form.Get("start"), 10, 64)
		end, _ := strconv.ParseInt(r.Form.Get("end"), 10, 64)
		logs.Info("TransResHandler, channel:", channel, "device:", device, "start:", start, "end:", end)
		if transRes := GetTransRes(channel, device, start, end); transRes != nil {
			fmt.Fprintln(w, util.ToJson(transRes))
		} else {
			fmt.Fprintln(w, "no result!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func GetSellCfgHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "GET" {
		r.ParseForm()
		if ver := r.Form.Get("version"); ver == "all" {
			fmt.Fprintln(w, util.ToJson(GetSellCfg2(true, "")))
		} else if ver == "" || ver == "0" {
			type CfgState struct { //照顾前端逻辑临时加的结构
				Cfg      *SellCfgManager `json:"cfg"`
				AwardIds []string        `json:"award_ids"`
			}
			var cs CfgState
			cs.Cfg = GetSellCfg2(false, r.Form.Get("channel"))
			if loginname := r.Form.Get("loginname"); loginname != "" {
				cs.AwardIds = GetAwardIds(loginname)
				fmt.Fprintln(w, util.ToJson(cs))
			} else {
				fmt.Fprintln(w, util.ToJson(cs.Cfg))
			}
		} else {
			fmt.Fprintln(w, GetSellCfgVer())
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func AddSellCfgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		cfg := &SellMiniGameCfg{}
		cfg.MiniAppId = r.PostFormValue("mini_app_id")
		cfg.MiniGameId = r.PostFormValue("mini_game_id")
		cfg.MiniGamePath = r.PostFormValue("mini_game_path")
		cfg.MiniGameName = r.PostFormValue("mini_game_name")
		cfg.MiniGameDetail = r.PostFormValue("mini_game_detail")
		iconIndex, _ := strconv.ParseInt(r.PostFormValue("icon_index"), 10, 16)
		cfg.IconIndex = uint16(iconIndex)
		cfg.Channels = r.PostFormValue("channels")
		showPos, _ := strconv.ParseInt(r.PostFormValue("show_pos"), 10, 64)
		cfg.ShowPos = uint64(showPos)
		promoteLevel, _ := strconv.ParseInt(r.PostFormValue("promote_level"), 10, 8)
		cfg.PromoteLevel = uint8(promoteLevel)
		sortLevel, _ := strconv.ParseInt(r.PostFormValue("sort_level"), 10, 16)
		cfg.SortLevel = uint16(sortLevel)
		stayTime, _ := strconv.ParseInt(r.PostFormValue("stay_time"), 10, 16)
		cfg.StayTime = uint16(stayTime)
		award, _ := strconv.ParseInt(r.PostFormValue("award"), 10, 32)
		cfg.Award = uint32(award)
		logs.Info("AddSellCfgHandler, cfg:", cfg)
		if AddSellCfg(cfg) {
			fmt.Fprintln(w, "success")
		} else {
			fmt.Fprintln(w, "[ERROR]add mini game config failed!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func ModSellCfgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		flag := uint32(0)
		cfg := &SellMiniGameCfg{}
		cfg.MiniAppId = r.PostFormValue("mini_app_id")
		cfg.MiniGameId = r.PostFormValue("mini_game_id")
		if vs := r.PostForm["mini_game_path"]; len(vs) > 0 {
			cfg.MiniGamePath = vs[0]
			flag |= FLAG_MINI_GAME_PATH
		}
		if vs := r.PostForm["mini_game_name"]; len(vs) > 0 {
			cfg.MiniGameName = vs[0]
			flag |= FLAG_MINI_GAME_NAME
		}
		if vs := r.PostForm["mini_game_detail"]; len(vs) > 0 {
			cfg.MiniGameDetail = vs[0]
			flag |= FLAG_MINI_GAME_DETAIL
		}
		if vs := r.PostForm["icon_index"]; len(vs) > 0 {
			iconIndex, _ := strconv.ParseInt(vs[0], 10, 16)
			cfg.IconIndex = uint16(iconIndex)
			flag |= FLAG_ICON_INDEX
		}
		if vs := r.PostForm["channels"]; len(vs) > 0 {
			cfg.Channels = vs[0]
			flag |= FLAG_CHANNELS
		}
		if vs := r.PostForm["show_pos"]; len(vs) > 0 {
			showPos, _ := strconv.ParseInt(vs[0], 10, 64)
			cfg.ShowPos = uint64(showPos)
			flag |= FLAG_SHOW_POS
		}
		if vs := r.PostForm["promote_level"]; len(vs) > 0 {
			promoteLevel, _ := strconv.ParseInt(vs[0], 10, 8)
			cfg.PromoteLevel = uint8(promoteLevel)
			flag |= FLAG_PROMOTE_LEVEL
		}
		if vs := r.PostForm["sort_level"]; len(vs) > 0 {
			sortLevel, _ := strconv.ParseInt(vs[0], 10, 16)
			cfg.SortLevel = uint16(sortLevel)
			flag |= FLAG_SORT_LEVEL
		}
		if vs := r.PostForm["stay_time"]; len(vs) > 0 {
			stayTime, _ := strconv.ParseInt(vs[0], 10, 16)
			cfg.StayTime = uint16(stayTime)
			flag |= FLAG_STAY_TIME
		}
		if vs := r.PostForm["award"]; len(vs) > 0 {
			award, _ := strconv.ParseInt(vs[0], 10, 32)
			cfg.Award = uint32(award)
			flag |= FLAG_AWARD
		}
		logs.Info("ModSellCfgHandler, cfg:", cfg, "flag:", flag)
		if flag == 0 || ModSellCfg(cfg, flag) {
			fmt.Fprintln(w, "success")
		} else {
			fmt.Fprintln(w, "[ERROR]update mini game config failed!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func DelSellCfgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		miniAppId := r.PostFormValue("mini_app_id")
		miniGameId := r.PostFormValue("mini_game_id")
		logs.Info("DelSellCfgHandler, miniAppId:", miniAppId, "miniGameId:", miniGameId)
		if DelSellCfg(miniAppId, miniGameId) {
			fmt.Fprintln(w, "success")
		} else {
			fmt.Fprintln(w, "[ERROR]delete mini game config failed!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	const maxSize = 10 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		fmt.Fprintln(w, "[ERROR]file too big!")
		return
	}
	miniAppId := r.PostFormValue("mini_app_id")
	miniGameId := r.PostFormValue("mini_game_id")
	if GetSellCfgById(miniAppId, miniGameId) == nil {
		fmt.Fprintln(w, "[ERROR]mini_app_id or mini_game_id invalid!")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, "[ERROR]"+err.Error())
		return
	}
	defer file.Close()
	filename := fmt.Sprintf("./download/game_%s%s.png", miniAppId, miniGameId)
	err = util.SaveToFile(file, filename)
	if err != nil {
		fmt.Fprintln(w, "[ERROR]"+err.Error())
		return
	}
	logs.Info("UploadHandler, miniGameId:", miniGameId, "filename:", filename)
	fmt.Fprintf(w, UploadIcon(miniAppId, miniGameId))
}

func SellEventHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		r.ParseForm()
		loginname := r.PostFormValue("loginname")
		channel := r.PostFormValue("channel")
		var os uint8 = 0
		if ua := useragent.Parse(r.Header.Get("User-Agent")); ua != nil {
			os = Device2OS(ua.OS)
		}
		miniAppId := r.PostFormValue("mini_app_id")
		miniGameId := r.PostFormValue("mini_game_id")
		event, _ := strconv.ParseInt(r.PostFormValue("event"), 10, 16)
		logs.Info("SellEventHandler, loginname:", loginname, "channel:", channel, "os:", os, "miniAppId:", miniAppId, "miniGameId:", miniGameId, "event:", event)
		AddSellEvent(loginname, channel, os, miniAppId, miniGameId, uint16(event))
		fmt.Fprintln(w, "success")
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func SellAwardIdsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		fmt.Fprintln(w, util.ToJson(GetAwardIds(r.Form.Get("loginname"))))
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func SellResHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		gameId, _ := strconv.ParseInt(r.Form.Get("game_id"), 10, 32)
		item := r.Form.Get("item")
		channel := r.Form.Get("channel")
		device := r.Form.Get("device")
		start, _ := strconv.ParseInt(r.Form.Get("start"), 10, 64)
		end, _ := strconv.ParseInt(r.Form.Get("end"), 10, 64)
		logs.Info("SellResHandler, gameId:", gameId, "item:", item, "channel:", channel, "device:", device, "start:", start, "end:", end)
		if lpRes := GetSellRes(int32(gameId), item, channel, device, start, end); lpRes != nil {
			fmt.Fprintln(w, util.ToJson(lpRes))
		} else {
			fmt.Fprintln(w, "no result!")
		}
	} else {
		fmt.Fprintln(w, r.Method, "not support!")
	}
}

func InitServer() {
	db.InitDB()
	version.Init()
	InitSellCfg()
	InitSellEvent()
}

func StartServer() {
	InitServer()
	logs.Notice("http server start...")
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/land_page", LandPageHandler)
	http.HandleFunc("/land_page_res", LandPageResHandler)
	http.HandleFunc("/trans_app/", TransAppHandler)
	http.HandleFunc("/trans_event/", TransEventHandler)
	http.HandleFunc("/trans_res", TransResHandler)
	http.HandleFunc("/sell_cfg_get", GetSellCfgHandler)
	http.HandleFunc("/sell_cfg_add", AddSellCfgHandler)
	http.HandleFunc("/sell_cfg_mod", ModSellCfgHandler)
	http.HandleFunc("/sell_cfg_del", DelSellCfgHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/sell_event", SellEventHandler)
	http.HandleFunc("/sell_award_ids", SellAwardIdsHandler)
	http.HandleFunc("/sell_res", SellResHandler)
	http.Handle("/html/", http.StripPrefix("/html", http.FileServer(http.Dir("../template"))))
	if config.Get().Listen.Plat != "" {
		util.HttpListen(config.Get().Listen.Plat)
	}
	if config.Get().Listen.PlatTLS != "" {
		util.HttpsListen(config.Get().Listen.PlatTLS, "cert.pem", "key.pem")
	}
}

func CloseServer() {
	SaveSellCfg()
	SaveSellEvent()
}
