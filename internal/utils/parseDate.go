package utils

import (
	"fmt"
	"time"
)

func ParseDate(dateString string) time.Time {
	layout := "2006-01-02 15:04"
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		panic(fmt.Sprintf("Ошибка при преобразовании строки во временной объект: %v", err))
	}
	return parsedTime
}
