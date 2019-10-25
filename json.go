package gomamiso

import (
	"encoding/json"
	"io/ioutil"
)

type Setting struct {
	Database Database `json:"database"`
	Device   Device   `json:"device`
}

type Database struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type Device struct {
	Name    string `json:"name"`
	Macaddr string `json:"macaddr"`
}

func ReadAndSetInfo() (*Setting, error) {
	bytes, err := ioutil.ReadFile("setting.json")
	if err != nil {
		return nil, err
	}

	var setting Setting
	if err = json.Unmarshal(bytes, &setting); err != nil {
		return nil, err
	}

	return &setting, nil
}
