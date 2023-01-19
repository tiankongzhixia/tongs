package tong

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"tongs/config"

	"github.com/go-redis/redis"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/mozillazg/go-pinyin"
	"go.uber.org/zap"
)

var (
	managers   = make([]*Tongs, 0)
	Redis      *redis.Client
	BloomRedis *redis.Client
	Config     config.Tongs
	UserAgents = make(map[string][]string)
	args       = pinyin.NewArgs()
	taskIdMap  = map[string]string{}
	Log        *zap.Logger
	Status     = map[string]int{
		"stop":     0,
		"stopping": 1,
		"running":  2,
		"err":      3,
	}
)

type Manager struct{}

func (m *Manager) Init() {
	//初始化执行初始化操作
	for _, t := range managers {
		for _, task := range t.Tasks {
			task.Init()
		}
	}
}

// AddTongs 添加Tongs
func (m *Manager) AddTongs(name string) *Tongs {
	tongs, err := m.FindTongs(name)
	if tongs == nil && err != nil {
		t := newTongs(name)
		managers = append(managers, t)
		return t
	}
	return nil
}

// FindTongs 根据名称查找
func (m *Manager) FindTongs(name string) (*Tongs, error) {
	return findTongs(name)
}

// FindTongs 根据名称查找
func (m *Manager) FindTask(tongsName, taskName string) (*Task, error) {
	return findTask(tongsName, taskName)
}

// AddTask 添加任务
func (m *Manager) AddTask(tongsName string, task *Task) error {
	tongs, err := m.FindTongs(tongsName)
	if err != nil {
		return err
	}
	if err := tongs.AddTask(task); err != nil {
		return err
	}
	return nil
}

// AddTaskOrNew 添加任务,Tongs不存在则新建
func (m *Manager) AddTaskOrNew(tongsName string, task *Task) error {
	tongs, err := m.FindTongs(tongsName)
	if err != nil && tongs == nil {
		tongs = m.AddTongs(tongsName)
	}
	if err := tongs.AddTask(task); err != nil {
		return err
	}
	return nil
}

// AddTasks 给指定的Tongs批量添加任务
func (m *Manager) AddTasks(tongsName string, tasks ...*Task) error {
	tongs, err := m.FindTongs(tongsName)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := tongs.AddTask(t); err != nil {
			return err
		}
	}
	return nil
}

// GetTongsName 获取所有的TongsName
func (m *Manager) GetTongsName() []string {
	var ns []string
	for _, t := range managers {
		ns = append(ns, t.Name)
	}
	return ns
}

func (m *Manager) Stop(name string) error {
	if tongs, err := m.FindTongs(name); err != nil {
		return err
	} else {
		tongs.Stop()
		return nil
	}
}

func newTongs(name string) *Tongs {
	return &Tongs{Name: name, Tasks: make([]*Task, 0), lock: new(sync.Mutex), Ctx: make(map[string]interface{})}
}
func getTaskId(groupName string, taskName string) string {
	key := groupName + ":" + taskName
	if taskIdMap[key] != "" {
		return taskIdMap[key]
	}
	gn := strings.Join(pinyin.LazyPinyin(groupName, args), "")
	tn := strings.Join(pinyin.LazyPinyin(taskName, args), "")
	id := gn + ":" + tn
	taskIdMap[key] = id
	return id
}
func initStore(t *Task) {
	if Config.Bloom.Open {
		t.store = &BloomStore{
			Id:        t.ID,
			TongsName: strings.Join(pinyin.LazyPinyin(t.tongs.Name, args), ""),
			Client:    BloomRedis,
			IsQueue:   t.IsQueue,
		}
	} else {
		t.store = &TongsStore{
			Id:        t.ID,
			TongsName: strings.Join(pinyin.LazyPinyin(t.tongs.Name, args), ""),
			Client:    Redis,
			IsQueue:   t.IsQueue,
		}
	}

	if t.IsQueue {
		q, err := queue.New(t.Thread, t.store)
		if err != nil {
			panic(fmt.Sprintf("任务【%s】ID:【%s】队列创建失败,error:%s", t.tongs.Name+":"+t.Name, t.ID, err.Error()))
		}
		t.queue = q
	} else {
		t.collector.SetStorage(t.store)
	}
}

func autoUserAgent(task *Task) {
	if !task.AutoUA {
		return
	}
	Log.Debug(fmt.Sprintf("任务【%s-%s】启动自动切换UA", task.tongs.Name, task.Name))
	task.collector.OnRequest(func(r *colly.Request) {
		ua := RandomUAWithType(task.UaType)
		r.Headers.Set("User-Agent", ua)
		r.Headers.Set("user-agent", ua)
	})
}

func autoDelay(task *Task) {
	rule := &colly.LimitRule{}
	if task.Delay != 0 {
		rule.Delay = time.Duration(task.Delay) * time.Second
	}
	if task.Domain != "0" {
		rule.DomainGlob = task.Domain
	}
	if task.AutoDelay {
		Log.Debug(fmt.Sprintf("任务【%s-%s】启动随机延迟", task.tongs.Name, task.Name))
		rule.RandomDelay = 1 * time.Minute
	}
	if task.Thread > 0 {
		Log.Debug(fmt.Sprintf("任务【%s-%s】设置启动线程【%s】", task.tongs.Name, task.Name, strconv.Itoa(task.Thread)))
		rule.Parallelism = task.Thread
	}
	task.collector.Limit(rule)
}
func findTongs(name string) (*Tongs, error) {
	for _, t := range managers {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, errors.New("组不存在")
}

func findTask(tongsName, taskName string) (*Task, error) {
	for _, t := range managers {
		if t.Name == tongsName {
			return t.findTaskWithName(taskName)
		}
	}
	return nil, errors.New("组不存在")
}
