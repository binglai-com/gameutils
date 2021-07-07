package binglaiapi

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/binglai-com/gameutils"
)

type ApiHandler struct {
	Token  string //用来存放api使用的token
	AppId  string //AppId
	AppKey string //AppKey
}

var (
	authdomain         = "http://175.24.153.177:6000/v1"
	DialTimeOut        = 30 * time.Second
	ReadWriteTimeOut   = 10 * time.Second
	defaulthttpsetting httplib.BeegoHTTPSettings
)

func init() {
	defaulthttpsetting = httplib.BeegoHTTPSettings{
		UserAgent:        "beegoServer",
		ConnectTimeout:   DialTimeOut,
		ReadWriteTimeout: ReadWriteTimeOut,
		Gzip:             true,
		DumpBody:         true,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   DialTimeOut,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

}

//初始化api handler
func InitApi(appid string, appkey string) (*ApiHandler, error) {
	var ret = &ApiHandler{AppId: appid, AppKey: appkey, Token: ""}
	if err := ret.Refresh(); err != nil {
		return nil, err
	} else {
		return ret, nil
	}
}

//刷新签名
func (api *ApiHandler) Refresh() error {
	var ts = time.Now().Unix()
	var sign = gameutils.Md5Str(fmt.Sprintf("%s%s%d", api.AppId, api.AppKey, ts))
	req := httplib.Get(authdomain+"/auth/").
		Setting(defaulthttpsetting).
		Param("AppId", api.AppId).
		Param("Sign", sign).
		Param("Ts", fmt.Sprintf("%d", ts))

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case 444:
			return fmt.Errorf("444 Invalid AppId")
		case 445:
			return fmt.Errorf("445 Sign Out of time")
		case 446:
			return fmt.Errorf("446 Invalid Sign")
		case 500:
			return fmt.Errorf("500 Server error")
		default:
			return fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		tokenstr, _ := req.String()
		api.Token = tokenstr
		return nil
	}
}

const (
	BodyType_None uint8 = 0 + iota
	BodyType_Bytes
	BodyType_String
	BodyType_Json
	BodyType_Xml
	BodyTYpe_Yaml
)

//获取请求响应
func (api *ApiHandler) Response(req *httplib.BeegoHTTPRequest, body interface{}, bodytype uint8) (*http.Response, error) {
	req.Header("Authorization", api.Token)
	// req.Header("Authorization", api.Token)

	if bodytype != BodyType_None && body != nil {
		switch bodytype {
		case BodyType_Bytes, BodyType_String:
			req.Body(body)
		case BodyType_Json:
			if _, err := req.JSONBody(body); err != nil {
				return nil, err
			}
		case BodyType_Xml:
			if _, err := req.XMLBody(body); err != nil {
				return nil, err
			}
		case BodyTYpe_Yaml:
			if _, err := req.YAMLBody(body); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unknow Body Type %d", bodytype)
		}
	}

	rsp, err := req.DoRequest()
	if err != nil {
		return rsp, err
	}

	if rsp.StatusCode == http.StatusUnauthorized { //未授权 刷新token
		if err := api.Refresh(); err != nil { //token刷新失败
			return rsp, err
		}

		//token刷新后
		req.Header("Authorization", api.Token)
		req.GetRequest().Body = nil
		if bodytype != BodyType_None && body != nil {
			switch bodytype {
			case BodyType_Bytes, BodyType_String:
				req.Body(body)
			case BodyType_Json:
				if _, err := req.JSONBody(body); err != nil {
					return nil, err
				}
			case BodyType_Xml:
				if _, err := req.XMLBody(body); err != nil {
					return nil, err
				}
			case BodyTYpe_Yaml:
				if _, err := req.YAMLBody(body); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("Unknow Body Type %d", bodytype)
			}
		}

		rsp, err = req.DoRequest()
		if err != nil {
			return rsp, err
		}

		if rsp.StatusCode == http.StatusUnauthorized { //授权失败
			return rsp, fmt.Errorf("401 授权失败")
		}
	}

	return rsp, nil
}

func (api *ApiHandler) GetBytes(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(reader)
	} else {
		return ioutil.ReadAll(resp.Body)
	}
}
