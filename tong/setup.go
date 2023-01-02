package tong

import "github.com/gocolly/colly/v2"

type Setup struct {
	Name                     string
	requestCallbacks         []colly.RequestCallback
	responseCallbacks        []colly.ResponseCallback
	responseHeadersCallbacks []colly.ResponseHeadersCallback
	errorCallbacks           []colly.ErrorCallback
	scrapedCallbacks         []colly.ScrapedCallback
}
