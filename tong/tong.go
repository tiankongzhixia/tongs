package tong

import (
	"errors"
	"github.com/gocolly/colly/v2"
	"sync"
)

type Tongs struct {
	Name      string
	Tasks     []*Task //组内任务
	Ctx       map[string]interface{}
	saveFuc   func(map[string]interface{})
	items     []interface{}
	itemCount int
	lock      *sync.Mutex
}

// NewTaskWithQueue 创建队列任务
func (t *Tongs) NewTaskWithQueue(name string, options ...colly.CollectorOption) *Task {
	id := getTaskId(t.Name, name)
	task := &Task{
		Name:    name,
		ID:      id,
		Status:  Status["stop"],
		IsQueue: true,
	}
	task.collector = newCollector(task, options...)
	t.AddTask(task)
	return task
}

// NewTask 创建普通任务
func (t *Tongs) NewTask(name string, options ...colly.CollectorOption) *Task {
	id := getTaskId(t.Name, name)
	task := &Task{
		Name:    name,
		ID:      id,
		Status:  Status["stop"],
		IsQueue: false,
	}
	task.collector = newCollector(task, options...)
	return task
}

func (t *Tongs) SetSaveFuc(fuc func(map[string]interface{})) {
	t.saveFuc = fuc
}

// Run 启动Tongs所有任务
func (t *Tongs) Run(url ...string) error {
	if len(t.Tasks) == 0 {
		return errors.New("当前Tongs内没有任务")
	}
	for i, t := range t.Tasks {
		if err := t.Run(url[i]); err != nil {
			return err
		}
	}
	return nil
}

// Stop 停止Tongs所有任务
func (t *Tongs) Stop() {
	for i := range t.Tasks {
		t.Tasks[i].Stop()
	}
}

// RunTask 启动任务
func (t *Tongs) RunTask(taskName string, url string) error {
	if task, err := t.findTaskWithName(taskName); err != nil {
		return err
	} else if err = task.Run(url); err != nil {
		return err
	}
	return nil
}

// StopTask 停止任务
func (t *Tongs) StopTask(taskName string) error {
	if task, err := t.findTaskWithName(taskName); err != nil {
		return err
	} else {
		task.Stop()
		return nil
	}
}

// AddURL 给指定的任务添加url
func (t *Tongs) AddURL(taskName string, url string) error {
	if task, err := t.findTaskWithName(taskName); err != nil {
		return err
	} else {
		return task.AddURL(url)
	}
}

func (t *Tongs) AddURLWithCtx(taskName string, url string, m map[string]interface{}) error {
	if task, err := t.findTaskWithName(taskName); err != nil {
		return err
	} else {
		return task.addRequest(url, m)
	}
}

// AddTask 添加任务
func (t *Tongs) AddTask(task *Task) error {
	task.ID = getTaskId(t.Name, task.Name)
	task.tongs = t
	ta, err := t.findTaskWithName(task.Name)
	if err != nil && ta == nil {
		t.Tasks = append(t.Tasks, task)
		return nil
	}
	return errors.New("任务名称在当前组内重复")
}

// 根据任务名称查找
func (t *Tongs) findTaskWithName(taskName string) (*Task, error) {
	for _, task := range t.Tasks {
		if task != nil && task.ID == getTaskId(t.Name, taskName) {
			return task, nil
		}
	}
	return nil, errors.New("没有该任务")
}

// Save 保存item
func (t *Tongs) Save(item map[string]interface{}) error {
	if t.saveFuc == nil {
		return errors.New("未设置保存方法")
	}
	if Config.Save.Open {
		t.lock.Lock()
		t.items = append(t.items, item)
		t.lock.Unlock()
	} else if Config.Save.Count {
		t.lock.Lock()
		t.itemCount++
		t.lock.Unlock()
	}
	t.saveFuc(item)
	return nil
}

// ItemSize 获取item数量
func (t *Tongs) ItemSize() int {
	if Config.Save.Open {
		return len(t.items)
	}
	return t.itemCount
}

// ClearItems 清除item
func (t *Tongs) ClearItems() {
	t.lock.Lock()
	t.items = make([]interface{}, 0)
	t.itemCount = 0
	t.lock.Unlock()
}
