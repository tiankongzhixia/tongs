package tong

import (
	"github.com/gocolly/colly/v2"
)

func newCollector(task *Task, options ...colly.CollectorOption) *colly.Collector {
	collector := colly.NewCollector(options...)
	return collector
}
