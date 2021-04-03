package odkhttp

import (
	"errors"
)

var defaultClient = OdkHttpClient{}

func init() {
	defaultClient.init()
}

// 设置header
func SetHeader(keyvalues ...string) (err error) {
	if len(keyvalues)%2 != 0 {
		err = errors.New("Header 参数未成对出现!")
		return
	}

	for index, value := range keyvalues {
		if index%2 == 0 {
			defaultClient.header[value] = keyvalues[index+1]
		}
	}
	return
}
func SetBaseUrl(baseUrl string) {
	defaultClient.BaseUrl = baseUrl
}
func SetContentType(contentType string) {
	defaultClient.contentType = contentType
}

// get full url
func Get(urladd string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.Get(urladd, params)
	return
}

// post full url bytes
func PostBytes(urladd string, bodybytes []byte) (resp []byte, err error) {
	resp, err = defaultClient.PostBytes(urladd, bodybytes)

	return
}

// post json
func PostJson(urladd string, params interface{}) (resp []byte, err error) {
	resp, err = defaultClient.PostJson(urladd, params)

	return
}

// put json
func PutJson(urladd string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.PutJson(urladd, params)

	return
}

// delete json
func DeleteJson(urladd string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.DeleteJson(urladd, params)

	return
}

// get full url
func PathGet(path string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.Get(defaultClient.BaseUrl+path, params)
	return
}

// post full url bytes
func PathPostBytes(path string, bodybytes []byte) (resp []byte, err error) {
	resp, err = defaultClient.PostBytes(defaultClient.BaseUrl+path, bodybytes)
	return
}

// post json
func PathPostJson(path string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.PostJson(defaultClient.BaseUrl+path, params)
	return
}

// put json
func PathPutJson(path string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.PutJson(defaultClient.BaseUrl+path, params)
	return
}

// delete json
func PathDeleteJson(path string, params map[string]interface{}) (resp []byte, err error) {
	resp, err = defaultClient.DeleteJson(defaultClient.BaseUrl+path, params)
	return
}
