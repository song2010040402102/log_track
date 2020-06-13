package version

import (
	"db"
	"github.com/astaxie/beego/logs"
)

func Init() {
	rows, err := db.GetMySql().Query("select type, version from version")
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		logs.Error("InitVersion, query failed with %s", err)
		return
	}
	g_mapVersion = make(map[uint16]uint64)
	for rows.Next() {
		var t uint16
		var ver uint64
		err = rows.Scan(&t, &ver)
		if err != nil {
			logs.Error("InitVersion, scan failed with %s", err)
			return
		}
		g_mapVersion[t] = ver
	}
}

func Get(t uint16) uint64 {
	return g_mapVersion[t]
}

func Set(t uint16, ver uint64) {
	g_mapVersion[t] = ver
	_, err := db.GetMySql().Exec("update version set version=? where type=?", ver, t)
	if err != nil {
		logs.Error("SetVersion, update failed with", err)
	}
}

var g_mapVersion map[uint16]uint64
