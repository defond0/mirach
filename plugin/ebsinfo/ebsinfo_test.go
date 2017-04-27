package ebsinfo

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

var testInfo = []Volume{
	Volume{
		ID:         "vol-awsid",
		Type:       "gp2",
		Size:       8,
		CreateTime: 1444000000,
		Encrypted:  true,
		SnapshotID: "snap-0909",
		Attachments: []Attachment{
			Attachment{
				AttachTime: 1444000000,
				Device:     "dev/xvda",
				InstanceID: "i-iamaninstanceidya",
				State:      "attached",
			},
		},
	},
	Volume{
		ID:         "vol-awsidother",
		Type:       "gp2",
		Size:       8,
		CreateTime: 1444000000,
		Encrypted:  true,
		SnapshotID: "snap-0919",
		Attachments: []Attachment{
			Attachment{
				AttachTime: 1444000000,
				Device:     "dev/xvda",
				InstanceID: "i-iamaninstanceidya",
				State:      "attached",
			},
		},
	},
}

func TestEBSInfoGetString(t *testing.T) {
	ogEBSInfo := new(EBSInfoGroup)
	MockEBSInfoGroup := &MockInfoGroup{
		mockGetInfo: func() {
			ogEBSInfo.Volumes = testInfo
		},
		mockString: ogEBSInfo.String,
	}
	newEBSInfo := new(EBSInfoGroup)
	if err := json.Unmarshal([]byte(MockEBSInfoGroup.String()), &newEBSInfo); err != nil {
		t.Error("can't unmarshall EBSInfo")
	}
	for i, vol := range newEBSInfo.Volumes {
		if !(vol.ID == ogEBSInfo.Volumes[i].ID) {
			t.Error("Volumes do not match")
		}
	}
}
