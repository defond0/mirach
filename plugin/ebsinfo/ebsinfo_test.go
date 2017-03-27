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
		CreateTime: "0000001", //its an old ebsvolume
		Encrypted:  true,
		SnapshotID: "snap-0909",
		Attachments: []Attachment{
			Attachment{
				AttachTime: "0000001", // it was attached a long time ago
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
		CreateTime: "0000001", //its an old ebsvolume
		Encrypted:  true,
		SnapshotID: "snap-0919",
		Attachments: []Attachment{
			Attachment{
				AttachTime: "0000001", // it was attached a long time ago
				Device:     "dev/xvda",
				InstanceID: "i-iamaninstanceidya",
				State:      "attached",
			},
		},
	},
}

func TestEbsInfoGetString(t *testing.T) {
	ogEbsInfo := new(EbsInfoGroup)
	MockEbsInfoGroup := &MockInfoGroup{
		mockGetInfo: func() {
			ogEbsInfo.Volumes = testInfo
		},
		mockString: ogEbsInfo.String,
	}
	newEbsInfo := new(EbsInfoGroup)
	if err := json.Unmarshal([]byte(MockEbsInfoGroup.String()), &newEbsInfo); err != nil {
		t.Error("can't unmarshall EbsInfo")
	}
	for i, vol := range newEbsInfo.Volumes {
		if !(vol.ID == ogEbsInfo.Volumes[i].ID) {
			t.Error("Volumes do not match")
		}
	}
}
