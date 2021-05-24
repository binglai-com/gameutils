package binglaiapi

import (
	"fmt"
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
	authdomain       = "http://175.24.153.177:6000/v1"
	DialTimeOut      = 30 * time.Second
	ReadWriteTimeOut = 10 * time.Second
)

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
		SetTimeout(DialTimeOut, ReadWriteTimeOut).
		Param("AppId", api.AppId).
		Param("Sign", sign).
		Param("Ts", fmt.Sprintf("%d", ts))

	rsp, err := req.Response()
	if err != nil {
		return err
	}

	if rsp.StatusCode != 200 {
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

//获取请求响应
func (api *ApiHandler) Response(req *httplib.BeegoHTTPRequest) (*http.Response, error) {
	req.SetTimeout(DialTimeOut, ReadWriteTimeOut).
		Header("Authorization", api.Token)
	// req.Header("Authorization", api.Token)

	rsp, err := req.DoRequest()
	if err != nil {
		return rsp, err
	}

	if rsp.StatusCode == 401 { //未授权 刷新token
		if err := api.Refresh(); err != nil { //token刷新失败
			return rsp, err
		}

		//token刷新后
		req.Header("Authorization", api.Token)
		rsp, err = req.DoRequest()
		if err != nil {
			return rsp, err
		}

		if rsp.StatusCode == 401 { //授权失败
			return rsp, fmt.Errorf("401 授权失败")
		}
	}

	return rsp, nil
}
