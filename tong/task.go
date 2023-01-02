package tong

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	whatwgUrl "github.com/nlnwa/whatwg-url/url"
	"net/url"
)

var urlParser = whatwgUrl.NewParser(whatwgUrl.WithPercentEncodeSinglePercentSign())

type Task struct {
	tongs     *Tongs           `json:"-"`
	Name      string           `json:"name,omitempty"` //任务名称
	ID        string           `json:"ID,omitempty"`
	StartUrl  string           `json:"startUrl,omitempty"` //首次启动url，队列模式可为空，非队列必填
	Status    int              `json:"status,omitempty"`   //当前任务状态
	AutoUA    bool             `json:"autoUA"`
	AutoDelay bool             `json:"autoDelay"`
	Delay     int              `json:"delay"`
	MaxDepth  int              `json:"maxDepth"`
	Thread    int              `json:"thread"`
	UaType    string           `json:"uaType"`
	IsQueue   bool             `json:"isQueue"`
	Ctx       *colly.Context   `json:"-"`
	Domain    string           `json:"domain"`
	queue     *queue.Queue     `json:"-"` //任务队列
	collector *colly.Collector `json:"-"` //colly scraper job
	store     Store            `json:"-"` //存储器
}

func (t *Task) Init() {
	initStore(t)
	config := &Config
	t.AutoUA = config.AutoUa
	t.AutoDelay = config.AutoDelay
	t.MaxDepth = config.MaxDepth
	t.collector.MaxDepth = t.MaxDepth
	autoUserAgent(t)
	autoDelay(t)
}

// SetUaType SetStartUrl SetQueue SetCollector 建造者模式
func (t *Task) SetUaType(ut string) *Task {
	t.UaType = ut
	return t
}
func (t *Task) SetMaxDepth(md int) *Task {
	t.MaxDepth = md
	return t
}
func (t *Task) SetStartUrl(startUrl string) *Task {
	t.StartUrl = startUrl
	return t
}
func (t *Task) SetQueue(q *queue.Queue) *Task {
	t.queue = q
	return t
}
func (t *Task) SetDelay(delay int) *Task {
	t.Delay = delay
	return t
}
func (t *Task) SetDomain(domain string) *Task {
	t.Domain = domain
	return t
}
func (t *Task) SetThread(thread int) *Task {
	t.Thread = thread
	return t
}
func (t *Task) SetCollector(f func(*colly.Collector, *Task)) *Task {
	f(t.collector, t)
	return t
}

// AddURL 给当前任务添加url
func (t *Task) AddURL(URL string) error {
	if t.IsQueue {
		if t.Status == Status["running"] {
			Log.Debug(fmt.Sprintf("队列任务【%s-%s】追加请求: %s", t.tongs.Name, t.Name, URL))
			t.queue.AddURL(URL)
		} else {
			return t.Run(URL)
		}
	} else {
		Log.Debug(fmt.Sprintf("普通任务【%s-%s】追加请求: %s", t.tongs.Name, t.Name, URL))
		return t.collector.Visit(URL)
	}
	return nil
}

func (t *Task) addPostRequest(URL string, requestData map[string]string, ctx map[string]interface{}) error {
	newCtx := colly.NewContext()
	if ctx == nil {
		if t.Ctx != nil {
			t.Ctx.ForEach(func(k string, v interface{}) interface{} {
				newCtx.Put(k, v)
				return nil
			})
		}
	} else {
		for k := range ctx {
			newCtx.Put(k, ctx[k])
		}
	}
	if t.IsQueue {
		if t.Status == Status["running"] {
			u, err := urlParser.Parse(URL)
			if err != nil {
				return err
			}
			u2, err := url.Parse(u.Href(false))
			if err != nil {
				return err
			}
			bys, _ := json.Marshal(requestData)
			r := &colly.Request{
				URL:     u2,
				Method:  "POST",
				Ctx:     newCtx,
				Headers: nil,
				Body:    bytes.NewReader(bys),
			}
			Log.Debug(fmt.Sprintf("队列任务【%s-%s】追加请求并传递上下文: %s", t.tongs.Name, t.Name, URL))
			return t.queue.AddRequest(r)
		} else {
			return t.Run(URL)
		}
	} else {
		return t.collector.Post(URL, requestData)
	}
}

// 添加请求到colly, 复制当前上下文信息到下一个请求
func (t *Task) addRequest(URL string, m map[string]interface{}) error {
	newCtx := colly.NewContext()
	if m == nil {
		if t.Ctx != nil {
			t.Ctx.ForEach(func(k string, v interface{}) interface{} {
				newCtx.Put(k, v)
				return nil
			})
		}
	} else {
		for k := range m {
			newCtx.Put(k, m[k])
		}
	}
	if t.IsQueue {
		if t.Status == Status["running"] {
			u, err := urlParser.Parse(URL)
			if err != nil {
				return err
			}
			u2, err := url.Parse(u.Href(false))
			if err != nil {
				return err
			}
			r := &colly.Request{
				URL:    u2,
				Method: "GET",
				Ctx:    newCtx,
			}
			Log.Debug(fmt.Sprintf("队列任务【%s-%s】追加请求并传递上下文: %s", t.tongs.Name, t.Name, URL))
			t.queue.AddRequest(r)
		} else {
			return t.Run(URL)
		}
	} else {
		return t.collector.Request("GET", URL, nil, newCtx, nil)
	}
	return nil
}

// AddURLToTong 向另外一个tong的task添加url
func (t *Task) AddURLToTong(tongsName, taskName string, url string) error {
	if tong, err := findTongs(tongsName); err != nil {
		return err
	} else {
		return tong.AddURL(taskName, url)
	}
}

// AddURLToTongWith 向另外一个tong的task添加url并且附带上下文
func (t *Task) AddURLToTongWith(tongsName, taskName string, url string, m map[string]interface{}) error {
	if tong, err := findTongs(tongsName); err != nil {
		return err

	} else {
		return tong.AddURLWithCtx(taskName, url, m)
	}
}

// AddURLTo 向指定任务添加请求url
func (t *Task) AddURLTo(taskName string, url string) error {
	return t.AddURLToWith(taskName, url, nil)
}

// AddURLToWith 向指定任务添加请求url并追加传递的上下文内容
func (t *Task) AddURLToWith(taskName string, url string, m map[string]interface{}) error {
	err := t.tongs.AddURLWithCtx(taskName, url, m)
	return err
}

// AddCtx 添加上下文内容
func (t *Task) AddCtx(key string, value interface{}) {
	t.Ctx.Put(key, value)
}

// GetCtx 获取上下文内容
func (t *Task) GetCtx(key string) interface{} {
	return t.Ctx.Get(key)
}

// Run 启动任务
func (t *Task) Run(url string) error {
	if t.Status == Status["running"] {
		return nil
	}
	if t.IsQueue {
		return t.queueRun(url)
	} else {
		return t.collectorRun(url)
	}
}

// Stop 停止任务
func (t *Task) Stop() {
	t.Status = Status["stopping"]
	if t.IsQueue {
		t.queue.Stop()
		t.Status = Status["stop"]
		Log.Info(fmt.Sprintf("队列任务【%s-%s】已停止", t.tongs.Name, t.Name))
		return
	}
	t.collector.OnRequest(func(request *colly.Request) {
		if t.Status == Status["stopping"] {
			request.Abort()
			t.Status = Status["stop"]
			Log.Info(fmt.Sprintf("普通任务【%s-%s】已停止", t.tongs.Name, t.Name))
		}
	})
}

// Save 保存item
func (t *Task) Save(m map[string]interface{}) error {
	return t.tongs.Save(m)
}
func (t *Task) queueRun(url string) error {
	var startUrl string
	if url != "" {
		startUrl = url
	} else {
		startUrl = t.StartUrl
	}
	if startUrl != "" {
		if err := t.queue.AddURL(startUrl); err != nil {
			Log.Error(fmt.Sprintf("队列任务【%s-%s】添加startUrl失败, err:%s", t.tongs.Name, t.Name, err.Error()))
			t.Stop()
			return err
		}
	}
	t.Status = Status["running"]
	Log.Info(fmt.Sprintf("队列任务【%s-%s】启动", t.tongs.Name, t.Name))
	go func() {
		if err := t.queue.Run(t.collector); err != nil {
			Log.Error(fmt.Sprintf("队列任务【%s-%s】启动失败, err:%s", t.tongs.Name, t.Name, err.Error()))
			t.Stop()
		}
	}()
	return nil
}
func (t *Task) collectorRun(url string) error {
	var startUrl string
	if url != "" {
		startUrl = url
	} else {
		startUrl = t.StartUrl
	}
	if startUrl == "" {
		Log.Error(fmt.Sprintf("普通任务【%s-%s】启动失败, err:%s", t.tongs.Name, t.Name, "启动url必填"))
		t.Stop()
		return errors.New(fmt.Sprintf("【%s】普通任务启动url必填", t.Name))
	}
	t.Status = Status["running"]
	Log.Info(fmt.Sprintf("普通任务【%s-%s】启动", t.tongs.Name, t.Name))
	go func() {
		if err := t.collector.Visit(startUrl); err != nil {
			t.Stop()
		}
	}()
	return nil
}
