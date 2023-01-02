package initialize

import (
	"tongs/global"
	"tongs/tong"
)

func InitTongs() {
	tong.Config = global.CONFIG.Tongs
	tong.Log = global.Log
	for _, ua := range tong.Config.Ua {
		tong.UserAgents[ua.Label] = ua.Values
	}
	global.TongsManager.Init()
}
