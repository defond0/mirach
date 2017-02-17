// +build unit

package compinfo

import (
	"encoding/json"
	"testing"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/docker"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
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

func TestDockerString(t *testing.T) {
	ogD := new(Docker)
	mockD := &MockInfoGroup{
		mockGetInfo: func() {
			ogD.IDs = []string{"hash1", "hash2"}
			ogD.Stat = []docker.CgroupDockerStat{
				{
					ContainerID: "container1",
					Name:        "test container 1",
					Image:       "hash1",
					Status:      "Exited (0) forever ago",
					Running:     false,
				},
				{
					ContainerID: "container2",
					Name:        "test container 2",
					Image:       "hash2",
					Status:      "Up 2 hours",
					Running:     true,
				},
			}

		},
		mockString: ogD.String,
	}
	mockD.GetInfo()
	newD := new(Docker)
	if err := json.Unmarshal([]byte(mockD.String()), &newD); err != nil {
		t.Error("Not able to unmarshal into Docker")
	}
}

func TestLoadString(t *testing.T) {
	ogL := new(Load)
	mockL := &MockInfoGroup{
		mockGetInfo: func() {
			ogL.Avg = &load.AvgStat{
				Load1:  0.1,
				Load5:  0.5,
				Load15: 0.15,
			}
			ogL.Misc = &load.MiscStat{
				ProcsRunning: 4,
				ProcsBlocked: 1,
				Ctxt:         9,
			}
		},
		mockString: ogL.String,
	}
	mockL.GetInfo()
	newL := new(Load)
	if err := json.Unmarshal([]byte(mockL.String()), &newL); err != nil {
		t.Error("Not able to unmarshal into Load")
	}
}

func TestSysString(t *testing.T) {
	ogS := new(Sys)
	mockS := &MockInfoGroup{
		mockGetInfo: func() {
			ogS.Host = &host.InfoStat{
				Hostname:             "test-home",
				Uptime:               409042,
				BootTime:             1487324471,
				Procs:                2,
				OS:                   "linux",
				Platform:             "arch",
				PlatformVersion:      "",
				KernelVersion:        "4.8.11-1-ARCH",
				VirtualizationSystem: "vbox",
				VirtualizationRole:   "host",
				HostID:               "someHostHash",
			}
			ogS.CPUs = []cpu.InfoStat{
				{
					CPU:        0,
					VendorID:   "GenuineIntel",
					Family:     "6",
					Model:      "60",
					Stepping:   3,
					PhysicalID: "0",
					CoreID:     "0",
					Cores:      1,
					ModelName:  "Intel(R) Core(TM) i3-4000M CPU @ 2.40GHz",
					Mhz:        2394.458,
					CacheSize:  3072,
					Flags:      []string{"flags", "go", "here"},
				},
				{
					CPU:        1,
					VendorID:   "GenuineIntel",
					Family:     "6",
					Model:      "60",
					Stepping:   3,
					PhysicalID: "0",
					CoreID:     "1",
					Cores:      1,
					ModelName:  "Intel(R) Core(TM) i3-4000M CPU @ 2.40GHz",
					Mhz:        2394.458,
					CacheSize:  3072,
					Flags:      []string{"flags", "go", "here"},
				},
			}
		},
		mockString: ogS.String,
	}
	mockS.GetInfo()
	newS := new(Sys)
	if err := json.Unmarshal([]byte(mockS.String()), &newS); err != nil {
		t.Error("Not able to unmarshal into Sys")
	}
}
