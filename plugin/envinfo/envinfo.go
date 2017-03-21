package envinfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"gitlab.eng.cleardata.com/dash/mirach/plugin"

	jww "github.com/spf13/jwalterweatherman"
)

type EnvInfoGroup struct {
	CloudProvider     string            `json:"provider"`
	CloudProviderInfo map[string]string `json:"info"`
}

func hitAwsMagicIp(path string) (string, error) {
	res, err := http.Get(fmt.Sprintf("http://169.254.169.254/latest/meta-data/%s", path))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 { // OK
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return string(bodyBytes), err
	}
	return "", err
}

// IAmInAws returns a bool, is mirach in aws?
func IAmInAws() (bool, error) {
	id, err := hitAwsMagicIp("instance-id")
	if err != nil {
		return false, err
	}
	// this will match instances ids we could force them to be i-(string of exeactly 17 characters)
	// if they get longer like they have had a history of doing this will continue to work, essentially
	// saying that they must be at least as long as they are currently.
	matched, err := regexp.MatchString("^i-[[:alnum:]]{8,}$", id)
	if err != nil {
		return false, err
	}
	return matched, nil
}

func (e *EnvInfoGroup) getNullInfo() error {
	e.CloudProvider = "unknown"
	e.CloudProviderInfo = map[string]string{}
	return nil
}

func (e *EnvInfoGroup) getAwsInfo() error {
	instId, err := hitAwsMagicIp("instance-id")
	if err != nil {
		return err
	}
	instType, err := hitAwsMagicIp("instance-type")
	if err != nil {
		return err
	}
	e.CloudProvider = "aws"
	e.CloudProviderInfo = map[string]string{}
	e.CloudProviderInfo["instance-id"] = instId
	e.CloudProviderInfo["instance-type"] = instType
	return nil
}

func (e *EnvInfoGroup) String() string {
	s, _ := json.Marshal(e)
	return string(s)
}

func (e *EnvInfoGroup) GetInfo() {
	if ans, _ := IAmInAws(); ans {
		e.getAwsInfo()
		jww.DEBUG.Println("detected aws environment")
	} else {
		e.getNullInfo()
		jww.DEBUG.Println("detected unknown environment")
	}
}

func GetInfo() plugin.InfoGroup {
	info := new(EnvInfoGroup)
	info.GetInfo()
	return info
}

func String() string {
	info := new(EnvInfoGroup)
	info.GetInfo()
	return info.String()
}
