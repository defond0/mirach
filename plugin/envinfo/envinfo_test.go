package envinfo

import (
	"encoding/json"
	"testing"
)

type MockInfoGroup struct {
	mockGetInfo func()
	mockString  func() string
}

func (m *MockInfoGroup) GetInfo() {
	m.mockGetInfo()
}

func (m *MockInfoGroup) String() string {
	return m.mockString()
}

var testInfo = map[string]string{
	"instance-id":   "i-aminaws",
	"instance-type": "m4-large",
}

func TestEnvInfoGetString(t *testing.T) {
	ogEnvInfo := new(EnvInfoGroup)
	mockEnvInfo := &MockInfoGroup{
		mockGetInfo: func() {
			ogEnvInfo.CloudProvider = "aws"
			ogEnvInfo.CloudProviderInfo = testInfo
		},
		mockString: ogEnvInfo.String,
	}
	mockEnvInfo.GetInfo()
	newEnvInfo := new(EnvInfoGroup)
	if err := json.Unmarshal([]byte(mockEnvInfo.String()), &newEnvInfo); err != nil {
		t.Error("can't unmarshall EnvInfo")
	}
	if !(newEnvInfo.CloudProvider == ogEnvInfo.CloudProvider) {
		t.Error("cloudproviders do not match")
	}
	if !(newEnvInfo.CloudProviderInfo["instance-id"] == ogEnvInfo.CloudProviderInfo["instance-id"] && newEnvInfo.CloudProviderInfo["instance-type"] == ogEnvInfo.CloudProviderInfo["instance-type"]) {
		t.Error("cloudproviderinfo does not match")
	}
}
