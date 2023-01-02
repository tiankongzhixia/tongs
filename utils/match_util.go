package utils

import (
	"regexp"
	"strings"
)

func RegexMatch(key1 string, key2 string) bool {
	res, err := regexp.MatchString(key2, key1)
	if err != nil {
		panic(err)
	}
	return res
}

func KeyMatch(key1 string, key2 string) bool {
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	re := regexp.MustCompile(`:[^/]+`)
	key2 = re.ReplaceAllString(key2, "$1[^/]+$2")

	return RegexMatch(key1, "^"+key2+"$")
}

// ArraysIndexOf 获取数组中的位置
func ArraysIndexOf[T comparable](arrays []T, val T) int {
	index := -1
	for i, v := range arrays {
		if v == val {
			index = i
		}
	}
	return index
}

// ArraysContain 数组是否包含
func ArraysContain[T comparable](arrays []T, val T) bool {
	return ArraysIndexOf(arrays, val) > -1
}
