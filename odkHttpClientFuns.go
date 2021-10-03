package odkhttp

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"net/url"

	"time"

	"os"

	"net/http"
	"net/http/cookiejar"

	"strings"

	"github.com/mgr9525/go-ruisutil"
)

var globalTransport *http.Transport = &http.Transport{MaxConnsPerHost: 100}

// //////////////////////公有函数
// 设置header
func (ohc *OdkHttpClient) SetHeader(keyvalues ...string) (err error) {
	if len(keyvalues)%2 != 0 {
		err = errors.New("Header 参数未成对出现!")
		return
	}
	ohc.mutx.Lock()
	for index, value := range keyvalues {
		if index%2 == 0 {
			ohc.header[value] = keyvalues[index+1]
		}
	}
	ohc.mutx.Unlock()
	return
}
func (ohc *OdkHttpClient) SetBaseUrl(baseUrl string) {
	ohc.BaseUrl = baseUrl
}
func (ohc *OdkHttpClient) SetContentType(contentType string) {
	ohc.contentType = contentType
}

// ///////////////////私有函数
// 初始化client
func (ohc *OdkHttpClient) init() {
	ohc.client = &http.Client{Transport: globalTransport}
	jar, _ := cookiejar.New(nil)
	ohc.header = make(map[string]string)
	ohc.client.Jar = jar
	ohc.contentType = APPLICATION_JSON

}

// 初始化client
func (ohc *OdkHttpClient) Init() {
	ohc.client = &http.Client{}
	jar, _ := cookiejar.New(nil)
	ohc.header = make(map[string]string)
	ohc.client.Jar = jar
	ohc.needSign = false
}

// 签名管理

func (ohc *OdkHttpClient) SetSignParams(publicKey, privateKey, replacePath string) {
	ohc.privateKey, ohc.publicKey, ohc.needSign, ohc.replacePath = privateKey, publicKey, true, replacePath
}

// 签名算法
func getParamSign(secretkey, timestamp, method, requestPath, body string) string {
	// encodeurl := url.QueryEscape(requestPath)
	rawStr := timestamp + method + requestPath
	if body != "" {
		rawStr = rawStr + body
	}
	h := hmac.New(sha256.New, []byte(secretkey))
	h.Write([]byte(rawStr))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// get full url
func (ohc *OdkHttpClient) Get(urladd string, params interface{}) (resp []byte, err error) {
	// //fmt.Println(urladd)
	if params != nil {
		mapParams, ok := params.(map[string]interface{})
		if !ok {
			err = errors.New("get params convert to map failed.")
			return
		}
		if len(mapParams) > 0 {
			temKeyValues := make([]string, len(mapParams))
			i := 0
			for key, value := range mapParams {
				temKeyValues[i] = fmt.Sprintf("%v=%v", key, value)
				i++
			}
			paramStr := strings.Join(temKeyValues, "&")
			urladd = fmt.Sprintf("%v?%v", urladd, paramStr)
		}
	}
	// //fmt.Println(urladd)
	req, err := http.NewRequest(http.MethodGet, urladd, nil)
	if err != nil {
		return
	}
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for k, v := range ohc.header {
		req.Header[k] = []string{v}
	}
	ohc.mutx.Unlock()

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodGet, urlpath, "")
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	req.Close = true
	//fmt.Println(req.Header, ohc.header)
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// get full url
func (ohc *OdkHttpClient) GetWithHeader(urladd string, params interface{}, header map[string]string) (resp []byte, err error) {
	// //fmt.Println(urladd)
	if params != nil {
		mapParams, ok := params.(map[string]interface{})
		if !ok {
			err = errors.New("get params convert to map failed.")
			return
		}
		if len(mapParams) > 0 {
			temKeyValues := make([]string, len(mapParams))
			i := 0
			for key, value := range mapParams {
				temKeyValues[i] = fmt.Sprintf("%v=%v", key, value)
				i++
			}
			paramStr := strings.Join(temKeyValues, "&")
			urladd = fmt.Sprintf("%v?%v", urladd, paramStr)
		}
	}
	// //fmt.Println(urladd)
	req, err := http.NewRequest(http.MethodGet, urladd, nil)
	if err != nil {
		return
	}
	//此处还可以写req.Header.Set("User-Agent", "myClient")

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodGet, urlpath, "")
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	for k, v := range header {
		req.Header[k] = []string{v}
	}
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

func (ohc *OdkHttpClient) GetSortParams(urladd string, paramkeyvalues ...string) (resp []byte, err error) {
	if len(paramkeyvalues)%2 != 0 {
		err = errors.New("参数个数不正确")
	}
	temKeyValues := make([]string, len(paramkeyvalues)/2)
	j := 0
	for i, param := range paramkeyvalues {
		if i%2 == 0 {
			temKeyValues[j] = fmt.Sprintf("%v=%v", param, paramkeyvalues[i+1])
			j++
		}
	}
	paramStr := strings.Join(temKeyValues, "&")
	urladd = fmt.Sprintf("%v?%v", urladd, paramStr)
	req, err := http.NewRequest(http.MethodGet, urladd, nil)
	// //fmt.Println(urladd)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for key, value := range ohc.header {
		req.Header[key] = []string{value}
	}
	ohc.mutx.Unlock()
	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodGet, urlpath, "")
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}

	}
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// post full url bytes
func (ohc *OdkHttpClient) PostBytes(urladd string, bodybytes []byte) (resp []byte, err error) {
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPost, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for k, v := range ohc.header {
		req.Header[k] = []string{v}
	}
	ohc.mutx.Unlock()

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPost, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}

	// req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-type", ohc.contentType)
	//fmt.Println(req.Header, ohc.header)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// post full url bytes
func (ohc *OdkHttpClient) PostBytesWithHeader(urladd string, bodybytes []byte, header map[string]string) (resp []byte, err error) {
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPost, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPost, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}

	for k, v := range header {
		req.Header[k] = []string{v}
	}

	// req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-type", ohc.contentType)
	//fmt.Println(req.Header, header)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// post json
func (ohc *OdkHttpClient) PostJson(urladd string, params interface{}) (resp []byte, err error) {
	if params == nil {
		resp, err = ohc.PostBytes(urladd, nil)
		return
	}
	bodybytes, err := json.Marshal(params)
	if err != nil {
		return
	}

	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPost, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for key, value := range ohc.header {
		req.Header[key] = []string{value}
	}
	ohc.mutx.Unlock()
	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPost, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}

	req.Header.Set("Content-type", ohc.contentType)
	//fmt.Println("PostJson", req.Header, ohc.header)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}
func (ohc *OdkHttpClient) PostJsonWithHeader(urladd string, params interface{}, header map[string]string) (resp []byte, err error) {
	if params == nil {
		resp, err = ohc.PostBytes(urladd, nil)
		return
	}
	bodybytes, err := json.Marshal(params)
	if err != nil {
		return
	}
	// //fmt.Println(string(bodybytes))
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPost, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPost, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	for k, v := range header {
		req.Header[k] = []string{v}
	}
	req.Header.Set("Content-type", ohc.contentType)
	//fmt.Println("PostJsonWithHeader", req.Header, header)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// put json
func (ohc *OdkHttpClient) PutJson(urladd string, params interface{}) (resp []byte, err error) {
	if params == nil {
		resp, err = ohc.PutBytes(urladd, nil)
		return
	}

	bodybytes, err := json.Marshal(params)
	if err != nil {
		return
	}
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPut, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for key, value := range ohc.header {
		req.Header[key] = []string{value}
	}
	ohc.mutx.Unlock()

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPut, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}

	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}

	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}
func (ohc *OdkHttpClient) PutBytes(urladd string, bodybytes []byte) (resp []byte, err error) {
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPut, urladd, body)

	ohc.mutx.Lock()
	for k, v := range ohc.header {
		req.Header[k] = []string{v}
	}
	ohc.mutx.Unlock()
	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPut, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	// req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}
func (ohc *OdkHttpClient) PutJsonWithHeader(urladd string, params interface{}, header map[string]string) (resp []byte, err error) {
	if params == nil {
		resp, err = ohc.PutBytes(urladd, nil)
		return
	}

	bodybytes, err := json.Marshal(params)
	if err != nil {
		return
	}
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPut, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPut, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	for k, v := range header {
		req.Header[k] = []string{v}
	}
	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}

	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}
func (ohc *OdkHttpClient) PutBytesWithHeader(urladd string, bodybytes []byte, header map[string]string) (resp []byte, err error) {
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodPut, urladd, body)

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodPut, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	for k, v := range header {
		req.Header[k] = []string{v}
	}
	// req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// delete json
func (ohc *OdkHttpClient) DeleteJson(urladd string, params interface{}) (resp []byte, err error) {
	if params == nil {
		resp, err = ohc.DeleteBytes(urladd, nil)
		return
	}

	bodybytes, err := json.Marshal(params)
	if err != nil {
		return
	}
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodDelete, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for k, v := range ohc.header {
		req.Header[k] = []string{v}
	}
	ohc.mutx.Unlock()
	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodDelete, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// post full url bytes
func (ohc *OdkHttpClient) DeleteBytes(urladd string, bodybytes []byte) (resp []byte, err error) {
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodDelete, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	ohc.mutx.Lock()
	for k, v := range ohc.header {
		req.Header[k] = []string{v}
	}
	ohc.mutx.Unlock()
	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodDelete, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	// req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// post full url bytes
func (ohc *OdkHttpClient) DeleteBytesWithHeader(urladd string, bodybytes []byte, header map[string]string) (resp []byte, err error) {
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodDelete, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodDelete, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	for k, v := range header {
		req.Header[k] = []string{v}
	}
	// req.Header.Set("Content-type", "application/json")
	req.Header.Set("Content-type", ohc.contentType)
	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

func (ohc *OdkHttpClient) DeleteJsonWithHeader(urladd string, params interface{}, header map[string]string) (resp []byte, err error) {
	if params == nil {
		resp, err = ohc.DeleteBytes(urladd, nil)
		return
	}

	bodybytes, err := json.Marshal(params)
	if err != nil {
		return
	}
	body := bytes.NewBuffer(bodybytes)
	req, err := http.NewRequest(http.MethodDelete, urladd, body)
	//此处还可以写req.Header.Set("User-Agent", "myClient")

	// 检测是否需要签名
	if ohc.needSign {
		u, _ := url.Parse(urladd)
		timestamp := fmt.Sprintf("%v", time.Now().Unix())
		urlpath := u.Path
		// if u.RawQuery != "" {
		// 	urlpath = u.Path + "?" + u.RawQuery
		// }
		if ohc.replacePath != "" {
			urlpath = strings.Replace(urlpath, ohc.replacePath, "", 1)
		}
		sign := getParamSign(ohc.privateKey, timestamp, http.MethodDelete, urlpath, string(bodybytes))
		req.Header["accessKey"] = []string{ohc.publicKey}
		req.Header["reqTime"] = []string{timestamp}
		req.Header["sign"] = []string{sign}
	}
	for k, v := range header {
		req.Header[k] = []string{v}
	}
	req.Header.Set("Content-type", ohc.contentType)

	req.Close = true
	response, err := ohc.client.Do(req)
	if err != nil {
		return
	}
	// urlp, err := url.Parse(urladd)
	// ohc.client.Jar.SetCookies(urlp, response.Cookies())
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

// get full url
func (ohc *OdkHttpClient) PathGet(path string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = ohc.Get(ohc.BaseUrl+path, params)
	return
}
func (ohc *OdkHttpClient) PathGetWithHeader(path string, params map[string]interface{}, header map[string]string) (resp []byte, err error) {
	resp, err = ohc.GetWithHeader(ohc.BaseUrl+path, params, header)
	return
}

func (ohc *OdkHttpClient) PathGetSortParams(path string, paramkeyvalues ...string) (resp []byte, err error) {
	resp, err = ohc.GetSortParams(ohc.BaseUrl+path, paramkeyvalues...)
	return
}

// post full url bytes
func (ohc *OdkHttpClient) PathPostBytes(path string, bodybytes []byte) (resp []byte, err error) {
	resp, err = ohc.PostBytes(ohc.BaseUrl+path, bodybytes)
	return
}
func (ohc *OdkHttpClient) PathPostBytesWithHeader(path string, bodybytes []byte, header map[string]string) (resp []byte, err error) {
	resp, err = ohc.PostBytesWithHeader(ohc.BaseUrl+path, bodybytes, header)
	return
}

// post json
func (ohc *OdkHttpClient) PathPostJson(path string, params interface{}) (resp []byte, err error) {
	resp, err = ohc.PostJson(ohc.BaseUrl+path, params)
	return
}
func (ohc *OdkHttpClient) PathPostJsonWithHeader(path string, params interface{}, header map[string]string) (resp []byte, err error) {
	resp, err = ohc.PostJsonWithHeader(ohc.BaseUrl+path, params, header)
	return
}

// put json
func (ohc *OdkHttpClient) PathPutJson(path string, params interface{}) (resp []byte, err error) {
	resp, err = ohc.PutJson(ohc.BaseUrl+path, params)
	return
}
func (ohc *OdkHttpClient) PathPutBytes(path string, bodybytes []byte) (resp []byte, err error) {

	resp, err = ohc.PutBytes(ohc.BaseUrl+path, bodybytes)
	return
}

// delete json
func (ohc *OdkHttpClient) PathDeleteJson(path string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = ohc.DeleteJson(ohc.BaseUrl+path, params)
	return
}

// delete json
func (ohc *OdkHttpClient) PathDeleteBytes(path string, bodybytes []byte) (resp []byte, err error) {
	resp, err = ohc.DeleteBytes(ohc.BaseUrl+path, bodybytes)
	return
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

// 上传文件
func (ohc *OdkHttpClient) UploadFile(urlAdd string, filePath string, progressValue chan FileUploadProgress) (resp []byte, err error) {
	body := ruisUtil.NewCircleByteBuffer(10240)
	boundary := randomBoundary()
	boundarybytes := []byte("\r\n--" + boundary + "\r\n")
	endbytes := []byte("\r\n--" + boundary + "--\r\n")

	reqest, err := http.NewRequest("POST", urlAdd, body)
	if err != nil {
		return
	}
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Content-Type", "multipart/form-data; charset=utf-8; boundary="+boundary)
	for key, value := range ohc.header {
		reqest.Header.Add(key, value)
	}
	go func() {
		//defer ruisRecovers("upload.run")
		f, err := os.OpenFile(filePath, os.O_RDONLY, 0666) //其实这里的 O_RDWR应该是 O_RDWR|O_CREATE，也就是文件不存在的情况下就建一个空文件，但是因为windows下还有BUG，如果使用这个O_CREATE，就会直接清空文件，所以这里就不用了这个标志，你自己事先建立好文件。
		if err != nil {
			return
		}
		stat, err := f.Stat() //获取文件状态
		if err != nil {
			return
		}
		defer f.Close()

		header := fmt.Sprintf("Content-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\nContent-Type: application/octet-stream\r\n\r\n", stat.Name())
		body.Write(boundarybytes)
		body.Write([]byte(header))

		fsz := (stat.Size())
		fupsz := int64(0)
		buf := make([]byte, 1024)
		progress := 0
		lastSize := int64(0)
		lastTime := time.Now().UnixNano()
		for {
			n, err := f.Read(buf)
			if n > 0 {
				nz, _ := body.Write(buf[0:n])
				fupsz += int64(nz)
				if int(float64(fupsz)/float64(fsz)*1000) > progress && time.Now().UnixNano()-1e9 > lastTime {
					progress = int(float64(fupsz) / float64(fsz) * 1000)
					speed := int64(fupsz-lastSize) * 1e9 / (time.Now().UnixNano() - lastTime)
					progressValue <- FileUploadProgress{
						ProgressValue: float64(fupsz) / float64(fsz) * 100,
						DoneSize:      fupsz,
						TotalSize:     fsz,
						Speed:         speed,
						NeedSeconds:   (fsz - fupsz) / speed,
					}
					lastSize = fupsz
					lastTime = time.Now().UnixNano()
				}

			}
			if err == io.EOF {
				break
			}
		}
		body.Write(endbytes)
		body.Write(nil) //输入EOF,表示数据写完了
	}()
	response, err := ohc.client.Do(reqest)
	if err != nil {
		return
	}
	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	return
}

type WriteCounter struct {
	Total     uint64
	lastTotal uint64
	lastTime  int64
	speed     float64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	nowTime := time.Now().UnixNano()
	if nowTime-wc.lastTime > 1e6 {
		wc.speed = float64(wc.Total-wc.lastTotal) * 1e6 / float64(nowTime-wc.lastTime)
		// Clear the line by using a character return to go back to the start and remove
		// the remaining characters by filling it with spaces
		fmt.Printf("\r%s", strings.Repeat(" ", 50))
		// Return again and print current status of download
		// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
		if wc.Total > 1000000000 {
			fmt.Printf("\r下载中... 已完成%.2f Gb 下载速度:%.2f Kb/s", float64(wc.Total)/1000000000.0, wc.speed)
		} else if wc.Total > 1000000 {
			fmt.Printf("\r下载中... 已完成%.2f Mb 下载速度:%.2f Kb/s", float64(wc.Total)/1000000.0, wc.speed)
		} else if wc.Total > 1000 {
			fmt.Printf("\r下载中... 已完成%.2f Kb 下载速度:%.2f Kb/s", float64(wc.Total)/1000.0, wc.speed)
		}

		wc.lastTime = nowTime
		wc.lastTotal = wc.Total
	}

}

// 下载文件
func (ohc *OdkHttpClient) DownloadFile(filePath string, urlAdd string) (err error) {
	// //fmt.Println("开始下载：", url)
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filePath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()
	req, err := http.NewRequest(http.MethodGet, urlAdd, nil)
	//此处还可以写req.Header.Set("User-Agent", "myClient")
	for key, value := range ohc.header {
		req.Header[key] = []string{value}
	}
	// Get the data
	resp, err := ohc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n下载完成,关闭文件\n")
	out.Close()

	err = os.Rename(filePath+".tmp", filePath)
	if err != nil {
		//fmt.Println("重命名失败", err)
		return err
	}

	return nil
}
