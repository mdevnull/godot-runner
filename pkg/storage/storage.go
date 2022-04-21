package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type (
	Env struct {
		Name      string `json:"name"`
		Scene     string `json:"scene"`
		Arguments string `json:"arguments"`
	}
	Project struct {
		ExecPath    string `json:"exec_path"`
		ProjectPath string `json:"project_path"`
		Envs        []Env  `json:"envs"`
	}
)

func New() *Project {
	return &Project{
		Envs: []Env{},
	}
}

func (p *Project) Save() error {
	buf, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("godot_runner.json", buf, os.ModePerm)
}

func (p *Project) TryLoad() error {
	buf, err := ioutil.ReadFile("godot_runner.json")
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, p)
}
