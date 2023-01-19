package baidu

import (
	"fmt"
	"tongs"
	"tongs/global"
	"tongs/tong"

	"github.com/gocolly/colly/v2"
)

var (
	TongsName = "百度"
	Task1     = "百度列表"
	Task2     = "获取相关搜索"
)

func init() {
	t := global.TongsManager.AddTongs(TongsName)
	t.SetSaveFuc(saveFunc) //添加保存方法，必须调用 tong.Task.Save() 方法才会执行
	//添加任务
	err := t.AddTask(t.NewTask(Task1).SetCollector(task1))
	if err != nil {
		panic(err.Error())
	}
	err = t.AddTask(t.NewTask(Task2).SetCollector(task2))
	if err != nil {
		panic(err.Error())
	}
}

var task1 = func(c *colly.Collector, task *tong.Task) {
	//相关搜索
	c.OnHTML("//*[@id=\"rs_new\"]/table/tbody/tr/td/a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			//添加url到`获取相关搜索`任务
			err := task.AddURLTo(Task2, link)
			if err != nil {
				global.Log.Error(fmt.Sprintf("添加到`获取相关搜索`任务失败 URL: %s ", link))
			}
		}

		//生成page页面地址，添加到当前任务继续执行
		for i := 0; i < 10; i++ {
			err := task.AddURL("http://nextUrl")
			if err != nil {
				global.Log.Error(fmt.Sprintf("添加到下一页失败 err: %s", err.Error()))
			}
		}

	})
}

var task2 = func(c *colly.Collector, task *tong.Task) {
	c.OnResponse(func(r *colly.Response) {
		global.Log.Info("请求地址：" + r.Request.URL.String())
		//获取到其他相关的路径，添加到任务1中
		task.AddURLTo(Task1, "http://****")

		//获取到其他相关的路径，添加到其他Tongs的任务中执行
		task.AddURLToTong("网易主页", "获取列表", "http://****")

		//... 具体操作
		item := tongs.M{
			"id":   "1",
			"name": "221",
		}
		//获取到其他相关的路径，添加到其他任务中执行 并附带上下文信息
		task.AddURLToWith(Task1, "http://****", item)

		//获取到其他相关的路径，添加到其他Tongs的任务中执行 并附带上下文信息
		task.AddURLToTongWith("网易主页", "获取列表", "http://****", item)

		//获取当前任务的上下文信息
		task.GetCtx("")

		if err := task.Save(item); err != nil {
			global.Log.Error(fmt.Sprintf("保存失败: %s ", err.Error()))
		}
	})
}

var saveFunc = func(item map[string]interface{}) {
	//...其他逻辑操作

	//global.DB.Save() 保存到数据库
	//global.Redis.Set("","",-1) //设置redis缓存等
}
