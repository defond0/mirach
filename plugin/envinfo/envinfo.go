package environment

import (
	"io/ioutil"
	"net/http"
	"regexp"
)

// InfoGroup is an interface for getting data and marshaling to json.
type InfoGroup interface {
	GetInfo()
	String() string
}

type EnvInfo struct {
	CloudProvider     string
	CloudProviderInfo map[string]string
}

//WhereAmI returns a string representing the name of the cloud provider, or err
// if they do not use a cloud provider in the following list:
//       aws
func WhereAmI() (string, error) {
	if ans, err := IAmInAws(); ans {
		return "aws", err
	}
	return "none", nil
}

// IAmInAws returns a bool, is mirach in aws?
func IAmInAws() (bool, error) {
	res, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		return false, nil
	}
	defer res.Body.Close()
	if res.StatusCode == 200 { // OK
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return false, err
		}
		// this will match instances ids we could force them to be i-(string of exeactly 17 characters)
		// if they get longer like they have had a history of doing this will continue to work, essentially
		// saying that they must be at least as long as they are currently.
		matched, err := regexp.MatchString("^i-[[:alnum:]]{8,}$", string(bodyBytes))
		return matched, err
	}
	return false, err

}
