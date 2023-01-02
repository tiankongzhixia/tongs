package tong

import "github.com/duke-git/lancet/v2/random"

// RandomUA 随机获取一个UA
func RandomUA() string {
	var uas []string
	for s := range UserAgents {
		uas = append(uas, UserAgents[s]...)
	}
	return uas[random.RandInt(0, len(uas)-1)]
}

// RandomUAWithType 根据类型随机获取一个UA
func RandomUAWithType(ut string) string {
	if ut == "" {
		return RandomUA()
	}
	uas := UserAgents[ut]
	return uas[random.RandInt(0, len(uas)-1)]
}
