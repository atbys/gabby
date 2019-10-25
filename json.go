package gomamiso

import (
	"encoding/json"
	"io/ioutil"
)

type SettingInfo struct {
	Database DatabaseInfo `json:"database"`
	Device   DeviceInfo   `json:"device`
}

type DatabaseInfo struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type DeviceInfo struct {
	Name    string `json:"name"`
	Macaddr string `json:"macaddr"`
}

func ReadAndSetInfo() (*SettingInfo, error) {
	bytes, err := ioutil.ReadFile("setting.json")
	if err != nil {
		return nil, err
	}

	var setting SettingInfo
	if err = json.Unmarshal(bytes, &setting); err != nil {
		return nil, err
	}

	return &setting, nil
}
