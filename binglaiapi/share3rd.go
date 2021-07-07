package binglaiapi

import (
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/client/httplib"
)

var (
	share3rddomain = "http://175.24.153.177:8080/v1"
)

//身份证两要素验证
func (api *ApiHandler) GetVerify(Id string, Name string) error {
	req := httplib.Get(share3rddomain+"/idverify/").
		Param("Id", Id).
		Param("Name", Name).
		Setting(defaulthttpsetting)

	rsp, err := api.Response(req, nil, 0)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK { //请求成功
		switch rsp.StatusCode {
		case 444:
			return fmt.Errorf("444 身份证号不合法")
		case 445:
			return fmt.Errorf("445 身份证号和姓名不一致")
		case 500:
			return fmt.Errorf("500 Server error")
		default:
			return fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	}

	return nil
}
