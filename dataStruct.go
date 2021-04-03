package odkhttp

import (
	"net/http"
	"sync"
)

const (
	APPLICATION_JSON          = "application/json"
	X_WWW_FORM_URLENCODED     = "application/x-www-form-urlencoded"
	WKJ_X_WWW_FORM_URLENCODED = "x-www-form-urlencoded"
	METHOD_GET                = "GET"
	METHOD_POST               = "POST"
	METHOD_DELETE             = "DELETE"
	METHOD_PUT                = "PUT"
)

// 基础连接结构体
type OdkHttpClient struct {
	BaseUrl     string
	client      *http.Client
	header      map[string]string
	contentType string
	WsProxyUrl  string
	mutx        sync.Mutex

	privateKey  string //用于签名
	publicKey   string //用于签名
	needSign    bool   //是否签名
	replacePath string //替换掉路径
}

type FileUploadProgress struct {
	ProgressValue float64
	DoneSize      int64
	TotalSize     int64
	Speed         int64
	NeedSeconds   int64
}
