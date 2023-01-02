package utils

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

func GetError(err error, r interface{}) string {
	s := reflect.TypeOf(r)
	errs := err.(validator.ValidationErrors)
	for _, fieldError := range errs {
		filed, _ := s.FieldByName(fieldError.Field())
		errTag := fieldError.Tag() + "_msg"
		// 获取对应binding得错误消息
		errTagText := filed.Tag.Get(errTag)
		// 获取统一错误消息
		errText := filed.Tag.Get("msg")
		if errTagText != "" {
			return errTagText
		}
		if errText != "" {
			return errText
		}
		return fieldError.Field() + ":" + fieldError.Tag()
	}
	return ""
}
