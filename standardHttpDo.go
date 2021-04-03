package odkhttp

import (
	"bytes"
	"errors"
	"io/ioutil"

	"net/http"
)

const (
	HTTP_GET    = "GET"
	HTTP_POST   = "POST"
	HTTP_PUT    = "PUT"
	HTTP_DELETE = "DELETE"

	HTTP_STATUS_WAIT_DONE  = 100
	HTTP_STATUS_DONE_OK    = 200
	HTTP_STATUS_DONE_ERROR = 500
)

type OdkHttpTask struct {
	TaskId      int64
	Header      map[string]string
	ContentType string
	Url         string
	PostData    []byte
	Method      string
	Resp        []byte
	Error       string
	Status      int
	Miner       string
	index       int
}

type OdkHttpTaskBook struct {
	taskBook heapAscSortById
	taskMap  map[int64]*OdkHttpTask
	capacity int
}

type heapAscSortById []*OdkHttpTask

// ID升序排列
func (arr heapAscSortById) swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
	arr[i].index, arr[j].index = i, j
}
func (arr heapAscSortById) less(i, j int) bool {
	return arr[i].TaskId < arr[j].TaskId
}
func (h heapAscSortById) up(i int) {
	for {
		f := (i - 1) / 2 // 父亲结点
		if i == f || h.less(f, i) {
			break
		}
		h.swap(f, i)
		i = f
	}
}
func (h heapAscSortById) down(i int) {
	for {
		l := 2*i + 1 // 左孩子
		if l >= len(h) {
			break // i已经是叶子结点了
		}
		j := l
		if r := l + 1; r < len(h) && h.less(r, l) {
			j = r // 右孩子
		}
		if h.less(i, j) {
			break // 如果父结点比孩子结点小，则不交换
		}
		h.swap(i, j) // 交换父结点和子结点
		i = j        //继续向下比较
	}
}
func (h *heapAscSortById) Push(x *OdkHttpTask) {
	x.index = len(*h)
	*h = append(*h, x)
	h.up(len(*h) - 1)
}
func (h *heapAscSortById) Remove(i int) (*OdkHttpTask, bool) {

	if i < 0 || i > len(*h)-1 {
		return nil, false
	}
	n := len(*h) - 1
	if i == len(*h)-1 {
		x := (*h)[n]
		*h = (*h)[0:n]
		return x, true
	}
	h.swap(i, n) // 用最后的元素值替换被删除元素
	// 删除最后的元素
	x := (*h)[n]
	*h = (*h)[0:n]
	// 如果当前元素大于父结点，向下筛选
	if h.less((i-1)/2, i) {
		h.down(i)
	} else { // 当前元素小于父结点，向上筛选
		h.up(i)
	}
	return x, true
}
func (h *heapAscSortById) Pop() *OdkHttpTask {
	n := len(*h) - 1
	h.swap(0, n)
	x := (*h)[n]
	*h = (*h)[0:n]
	h.down(0)
	return x
}
func (h heapAscSortById) Init() {
	n := len(h)
	// i > n/2-1 的结点为叶子结点本身已经是堆了
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}

func (mbob *OdkHttpTaskBook) Init(capacity int) {
	mbob.taskBook = make([]*OdkHttpTask, 0, capacity)
	mbob.taskMap = make(map[int64]*OdkHttpTask, capacity)
	mbob.capacity = capacity
}
func (mbob *OdkHttpTaskBook) Size() int { return len(mbob.taskBook) }

// 添加
func (mbob *OdkHttpTaskBook) Push(tasks ...*OdkHttpTask) {
	for _, task := range tasks {
		mbob.taskBook.Push(task)
		mbob.taskMap[task.TaskId] = task
	}
}

// 删除
func (mbob *OdkHttpTaskBook) Remove(taskId int64) (err error) {
	task, err := mbob.GetTask(taskId)
	if err != nil {
		return
	}
	task, ok := mbob.taskBook.Remove(task.index)
	if !ok {
		err = errors.New("ERROR:OdkHttpTaskBook 删除订单失败")
		return
	}
	delete(mbob.taskMap, taskId)
	return
}

// 获取任务
func (mbob *OdkHttpTaskBook) GetTask(taskId int64) (task *OdkHttpTask, err error) {
	task, ok := mbob.taskMap[taskId]
	if !ok {
		err = errors.New("ERROR:OdkHttpTaskBook 未发现订单")
		return
	}
	return
}

func (oht *OdkHttpTask) Do() (err error) {
	switch oht.Method {
	case HTTP_GET:
		req, nerr := http.NewRequest(http.MethodGet, oht.Url, nil)
		for key, value := range oht.Header {
			req.Header.Add(key, value)
		}
		req.Header.Set("Content-type", oht.ContentType)
		response, nerr := defaultClient.client.Do(req)
		err = nerr
		if err != nil {
			return
		}
		defer response.Body.Close()
		oht.Resp, err = ioutil.ReadAll(response.Body)
		if err != nil {
			oht.Error = err.Error()
			break
		}
	case HTTP_POST:
		req, nerr := http.NewRequest(http.MethodPost, oht.Url, bytes.NewBuffer(oht.PostData))
		for key, value := range oht.Header {
			req.Header.Add(key, value)
		}
		req.Header.Set("Content-type", oht.ContentType)
		response, nerr := defaultClient.client.Do(req)
		err = nerr
		if err != nil {
			break
		}
		defer response.Body.Close()
		oht.Resp, err = ioutil.ReadAll(response.Body)
		if err != nil {
			oht.Error = err.Error()
			break
		}
	case HTTP_PUT:
		req, nerr := http.NewRequest(http.MethodPut, oht.Url, bytes.NewBuffer(oht.PostData))
		for key, value := range oht.Header {
			req.Header.Add(key, value)
		}
		req.Header.Set("Content-type", oht.ContentType)
		response, nerr := defaultClient.client.Do(req)
		err = nerr
		if err != nil {
			break
		}
		defer response.Body.Close()
		oht.Resp, err = ioutil.ReadAll(response.Body)
		if err != nil {
			oht.Error = err.Error()
			break
		}
	case HTTP_DELETE:
		req, nerr := http.NewRequest(http.MethodDelete, oht.Url, bytes.NewBuffer(oht.PostData))
		for key, value := range oht.Header {
			req.Header.Add(key, value)
		}
		req.Header.Set("Content-type", oht.ContentType)
		response, nerr := defaultClient.client.Do(req)
		err = nerr
		if err != nil {
			break
		}
		defer response.Body.Close()
		oht.Resp, err = ioutil.ReadAll(response.Body)
		if err != nil {
			oht.Error = err.Error()
			break
		}
	default:
		err = errors.New("HTTP方法未识别" + oht.Method)
		oht.Error = err.Error()
	}
	if err != nil {
		oht.Error = err.Error()
		oht.Status = HTTP_STATUS_DONE_ERROR
		return
	}
	oht.Status = HTTP_STATUS_DONE_OK
	return
}
