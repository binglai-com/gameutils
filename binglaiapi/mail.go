package binglaiapi

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/beego/beego/v2/client/httplib"
)

var (
	maildomain = "http://175.24.153.177:6002/v1"

	maildomain_mailsends = []string{"http://175.24.153.177:6003/v1", "http://175.24.153.177:6004/v1"}

	// maildomain = "http://127.0.0.1:6002/v1"
)

func _getmailsendsdomain() string {
	return maildomain_mailsends[rand.Intn(len(maildomain_mailsends))]
}

const (
	CondType_Personal int = 0 + iota //个人邮件(发给单个玩家) 直接发送到个人身上，不记录到条件邮件列表中
	CondType_All                     //全区全服(所有角色)
	CondType_NameList                //多人邮件(按照名单依次发送)
	CondType_SvrList                 //按服务器接收(满足区服条件的玩家可以领取)
)

type MailItem struct {
	Id  int `json:"id"bson:"id"`   //物品id
	Num int `json:"num"bson:"num"` //物品数量
}

type Mail struct {
	Id       int64      `json:"id"bson:"_id"`            //邮件id 从0开始递增
	Ctime    int64      `json:"ct"bson:"ct"`             //邮件发送时间戳 单位 秒
	Title    string     `json:"title"bson:"title"`       //邮件标题
	Body     string     `json:"body"bson:"body"`         //邮件正文
	Type     int        `json:"type"bson:"type"`         //邮件类型、频道
	Items    []MailItem `json:"items"bson:"items"`       //邮件附件
	Source   string     `json:"source"bson:"source"`     //邮件来源备注
	CondType int        `json:"condtype"bson:"condtype"` //接收条件类型 (0个人邮件(发给单个玩家) 1全区全服(所有角色) 2多人邮件(按照名单依次发送) 3按服务器接收(满足区服条件的玩家可以领取))
	CondList []string   `json:"condlist"bson:"condlist"` //条件清单
}

//玩家邮箱中的邮件描述
type BoxMailDesc struct {
	Id        int64  `json:"id"`        //邮件id
	Ctime     int64  `json:"ct"`        //邮件发送的时间戳，单位 （秒）
	Title     string `json:"title"`     //邮件标题
	Type      int    `json:"type"`      //邮件类型
	IsRead    bool   `json:"isread"`    //阅读状态
	IsGet     bool   `json:"isget"`     //获取状态
	HaveItems bool   `json:"haveitems"` //有无可领附件
}

//玩家邮箱中的邮件详情
type BoxMail struct {
	Id     int64      `json:"id"bson:"id"`         //邮件id
	Ctime  int64      `json:"ct"bson:"ct"`         //邮件发送时间戳，单位（秒）
	Title  string     `json:"title"bson:"title"`   //邮件标题
	Body   string     `json:"body"bson:"body"`     //邮件正文
	Type   int        `json:"type"bson:"type"`     //邮件类型、频道
	Items  []MailItem `json:"items"bson:"items"`   //邮件附件
	IsRead bool       `json:"isread"bson:"isread"` //有无查阅
	IsGet  bool       `json:"isget"bson:"isget"`   //邮件附件是否已领取
}

//新增邮件
func (api *ApiHandler) CreateMail(Title string, Body string, Type int, CondType int, CondList []string, Items []MailItem, Source string) (*Mail, error) {
	var addmail = Mail{
		Title:    Title,
		Body:     Body,
		Type:     Type,
		Items:    Items,
		Source:   Source,
		CondType: CondType,
		CondList: CondList,
	}

	var req = httplib.Post(_getmailsendsdomain() + "/mails/")
	rsp, err := api.Response(req, addmail, BodyType_Json)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case 444:
			return nil, fmt.Errorf("444 无效的条件类型")
		case 445:
			return nil, fmt.Errorf("445 无效的条件列表")
		case 500:
			return nil, fmt.Errorf("500 Server error")
		default:
			return nil, fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		if res, err := api.GetBytes(rsp); err != nil {
			return nil, err
		} else {
			var retmail = new(Mail)
			if err := json.Unmarshal(res, retmail); err != nil {
				return nil, err
			} else {
				return retmail, nil
			}
		}
	}
}

//获取玩家邮件列表
func (api *ApiHandler) GetPlayerMailList(pid string) ([]BoxMailDesc, error) {
	var req = httplib.Get(maildomain + "/mailbox/pid/" + pid)
	rsp, err := api.Response(req, nil, 0)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case 500:
			return nil, fmt.Errorf("500 Server error")
		default:
			return nil, fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		if res, err := api.GetBytes(rsp); err != nil {
			return nil, err
		} else {
			var maildescs = []BoxMailDesc{}
			if err := json.Unmarshal(res, &maildescs); err != nil {
				return nil, err
			} else {
				return maildescs, nil
			}
		}
	}
}

//获取玩家邮箱中的邮件详情
func (api *ApiHandler) ReadPlayerMail(pid string, mailid int64) (*BoxMail, error) {
	var req = httplib.Get(maildomain + "/mailbox/pid/" + pid + "/mailid/" + fmt.Sprintf("%d", mailid))
	rsp, err := api.Response(req, nil, 0)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case 500:
			return nil, fmt.Errorf("500 Server error")
		default:
			return nil, fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		if res, err := api.GetBytes(rsp); err != nil {
			return nil, err
		} else {
			var maildetail = new(BoxMail)
			if err := json.Unmarshal(res, maildetail); err != nil {
				return nil, err
			} else {
				return maildetail, nil
			}
		}
	}
}

//领取一封邮件的邮件奖励
func (api *ApiHandler) ClaimMailItem(pid string, mailid int64) ([]MailItem, error) {
	var req = httplib.Put(maildomain + "/mailbox/pid/" + pid + "/mailid/" + fmt.Sprintf("%d", mailid))
	rsp, err := api.Response(req, nil, 0)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("404 Not Found")
		case 500:
			return nil, fmt.Errorf("500 Server error")
		default:
			return nil, fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		if res, err := api.GetBytes(rsp); err != nil {
			return nil, err
		} else {
			var getitems = []MailItem{}
			if err := json.Unmarshal(res, &getitems); err != nil {
				return nil, err
			} else {
				return getitems, nil
			}
		}
	}
}

//从玩家邮箱中删除一封邮件
func (api *ApiHandler) DeletePlayerMailById(pid string, mailid int64) error {
	var req = httplib.Delete(maildomain + "/mailbox/pid/" + pid + "/mailid/" + fmt.Sprintf("%d", mailid))
	rsp, err := api.Response(req, nil, 0)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("404 Not Found")
		case 500:
			return fmt.Errorf("500 Server error")
		default:
			return fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		return nil
	}
}

//一键领取所有可领取的邮件奖励
func (api *ApiHandler) ClaimAllMailItem(pid string) ([]MailItem, error) {
	var req = httplib.Post(maildomain + "/mailbox/batch/update/pid/" + pid)
	rsp, err := api.Response(req, nil, 0)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("404 Not Found")
		case 500:
			return nil, fmt.Errorf("500 Server error")
		default:
			return nil, fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		if res, err := api.GetBytes(rsp); err != nil {
			return nil, err
		} else {
			var getitems = []MailItem{}
			if err := json.Unmarshal(res, &getitems); err != nil {
				return nil, err
			} else {
				return getitems, nil
			}
		}
	}
}

//批量删除玩家邮件
func (api *ApiHandler) DeletePlayerMails(pid string, delmailids []int64) error {
	var req = httplib.Post(maildomain + "/mailbox/batch/delete/pid/" + pid)
	rsp, err := api.Response(req, delmailids, BodyType_Json)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case 500:
			return fmt.Errorf("500 Server error")
		default:
			return fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		return nil
	}
}

//向玩家邮箱中复制邮件 （仅用于旧邮件系统升级使用，正常发送邮件请使用 CreateMail ）
func (api *ApiHandler) PasteMails2MailBox(pid string, maildata BoxMail) (*BoxMail, error) {
	var req = httplib.Post(_getmailsendsdomain() + "/mailbox/pid/" + pid)
	if maildata.Id != 0 { //除非你知道如何正确的获取邮件id，否则你不应该使用明确的邮件id，可以预见到这样做将会产生重复的邮件id
		maildata.Id = 0
	}
	rsp, err := api.Response(req, maildata, BodyType_Json)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		switch rsp.StatusCode {
		case 500:
			return nil, fmt.Errorf("500 Server error")
		default:
			return nil, fmt.Errorf("%d Unknow Status Code", rsp.StatusCode)
		}
	} else {
		if res, err := api.GetBytes(rsp); err != nil {
			return nil, err
		} else {
			var retmail = new(BoxMail)
			if err := json.Unmarshal(res, retmail); err != nil {
				return nil, err
			} else {
				return retmail, nil
			}
		}
	}
}
