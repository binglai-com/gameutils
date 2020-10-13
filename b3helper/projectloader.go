package b3helper

import (
	"encoding/json"
	"gameutils/gamelog/filelog"
	"io/ioutil"

	"github.com/magicsea/behavior3go"
	"github.com/magicsea/behavior3go/config"
	"github.com/magicsea/behavior3go/core"
	"github.com/magicsea/behavior3go/loader"
)

//behavior3go 项目配置
type B3Project struct {
	Trees []config.BTTreeCfg `json:"trees"`
}

//根据title查找行为树配置
func (p *B3Project) FindTreeConfigByTitle(title string) *config.BTTreeCfg {
	for k, v := range p.Trees {
		if v.Title == title {
			return &p.Trees[k]
		}
	}
	return nil
}

//根据title和自定义节点map创建行为树
func (p *B3Project) CreateBevTreeWithTitle(title string, extMap *behavior3go.RegisterStructMaps) *core.BehaviorTree {
	bconf := p.FindTreeConfigByTitle(title)
	if bconf == nil {
		return nil
	}
	return loader.CreateBevTreeFromConfig(bconf, extMap)
}

//加载B3Editor导出的工程文件
func LoadB3ProjConfig(path string) (*B3Project, error) {
	var proj B3Project
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &proj)
	if err != nil {
		return nil, err
	}

	filelog.INFO("b3helper", "load b3 project from ", path, " succ, load tree cnt : ", len(proj.Trees))
	return &proj, nil
}
