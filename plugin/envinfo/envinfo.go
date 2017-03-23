package envinfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"gitlab.eng.cleardata.com/dash/mirach/plugin"

	jww "github.com/spf13/jwalterweatherman"
)

var Env *EnvInfoGroup

type EnvInfoGroup struct {
	CloudProvider     string            `json:"provider"`
	CloudProviderInfo map[string]string `json:"info"`
}

func hitAwsMagicIp(path string) ([]byte, error) {
	res, err := http.Get(fmt.Sprintf("http://169.254.169.254/latest/meta-data/%s", path))
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 { // OK
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return []byte{}, err
		}
		return bodyBytes, err
	}
	return []byte{}, err
}

// IAmInAws returns a bool, is mirach in aws?
func IAmInAws() (bool, error) {
	id, err := hitAwsMagicIp("instance-id")
	if err != nil {
		return false, err
	}
	// this will match instances ids we could force them to be i-(string of at least 8 characters)
	// if they get longer like they have had a history of doing this will continue to work, essentially
	// saying that they must be at least as long as they are currently.
	matched, err := regexp.MatchString("^i-[[:alnum:]]{8,}$", string(id))
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
	instID, err := hitAwsMagicIp("instance-id")
	if err != nil {
		return err
	}
	instType, err := hitAwsMagicIp("instance-type")
	if err != nil {
		return err
	}
	var iamInfoMap map[string]interface{}
	iamInfo, err := hitAwsMagicIp("iam/info")
	if err != nil {
		return err
	}
	err = json.Unmarshal(iamInfo, &iamInfoMap)
	if err != nil {
		return err
	}
	accountID := strings.Split(iamInfoMap["InstanceProfileArn"].(string), ":")[4]

	az, err := hitAwsMagicIp("placement/availability-zone")
	if err != nil {
		return err
	}
	regionExp := regexp.MustCompile("^[[:alpha:]]+-[[:alpha:]]+-[[:digit:]]+")
	region := regionExp.Find(az)
	e.CloudProvider = "aws"
	e.CloudProviderInfo = map[string]string{}
	e.CloudProviderInfo["instance-id"] = string(instID)
	e.CloudProviderInfo["account-id"] = string(accountID)
	e.CloudProviderInfo["instance-type"] = string(instType)
	e.CloudProviderInfo["availablity-zone"] = string(az)
	e.CloudProviderInfo["region"] = string(region)
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
