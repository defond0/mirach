package parsers

import (
	"fmt"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// SFALSE is returned by CoInitializeEx if it was already called on this thread.
// https://github.com/StackExchange/wmi/blob/master/wmi.go#L54
const SFALSE = 0x00000001

// GetWindowsKBs creates map of available, installed and available security kbs from Windows Update Agent as well as a list of errors that occurred generating that list.
func GetWindowsKBs() (map[string]map[string]KBArticle, []error) {
	err := coInit()
	defer ole.CoUninitialize()
	errors := []error{}
	out := make(map[string]map[string]KBArticle)
	avail, err := getWindowsAvailableKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["available"] = avail
	availSec, err := getWindowsAvailableSecurityKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["available_security"] = availSec
	installed, err := getWindowsInstalledKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["installed"] = installed
	return out, errors
}

func getWindowsInstalledKBs() (map[string]KBArticle, error) {
	art := map[string]KBArticle{}
	updates, err := searchUpdates("IsInstalled=1")
	if err != nil {
		return nil, err
	}
	defer updates.Release()
	var kbID string
	for update, length, err := updates.Next(1); length > 0; update, length, err = updates.Next(1) {
		if err != nil {
			return nil, err
		}
		defer update.Clear()
		updateDispatch := update.ToIDispatch()
		defer updateDispatch.Release()
		sev, err := updateDispatch.GetProperty("MsrcSeverity")
		kbs, err := updateDispatch.GetProperty("KBArticleIDs")
		if err != nil {
			return nil, err
		}
		kbIDs, err := getEnumFromDispatch(kbs.ToIDispatch())
		for kb, length, _ := kbIDs.Next(1); length > 0; kb, length, err = kbIDs.Next(1) {
			defer kb.Clear()
			newID := fmt.Sprintf("KB%s", kb.Value())
			if newID != kbID {
				kbID = newID
				security := sev.Value() == "Critical"
				art[kbID] = KBArticle{
					name:     kbID,
					Security: security,
				}
			}

		}
	}
	return art, nil
}

func getWindowsAvailableKBs() (map[string]KBArticle, error) {
	art := map[string]KBArticle{}
	updates, err := searchUpdates("IsInstalled=0")
	if err != nil {
		return nil, err
	}
	defer updates.Release()
	var kbID string
	for update, length, err := updates.Next(1); length > 0; update, length, err = updates.Next(1) {
		if err != nil {
			return nil, err
		}
		defer update.Clear()
		updateDispatch := update.ToIDispatch()
		defer updateDispatch.Release()
		sev, err := updateDispatch.GetProperty("MsrcSeverity")
		kbs, err := updateDispatch.GetProperty("KBArticleIDs")
		kbIDs, err := getEnumFromDispatch(kbs.ToIDispatch())
		if err != nil {
			return nil, err
		}

		for kb, length, _ := kbIDs.Next(1); length > 0; kb, length, err = kbIDs.Next(1) {
			newID := fmt.Sprintf("KB%s", kb.Value())
			if newID != kbID {
				kbID = newID
				security := sev.Value() == "Critical"
				art[kbID] = KBArticle{
					name:     kbID,
					Security: security,
				}
			}

		}

	}
	return art, nil
}

// https://msdn.microsoft.com/en-us/library/windows/desktop/aa386906(v=vs.85).aspx
func getWindowsAvailableSecurityKBs() (map[string]KBArticle, error) {
	art, err := getWindowsAvailableKBs()
	if err != nil {
		return nil, err
	}

	artSec := map[string]KBArticle{}
	for k, v := range art {
		if v.Security {
			artSec[k] = v
		}
	}
	return artSec, nil
}

func coInit() error {
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != SFALSE {
			return err
		}
	}
	return nil
}

//seek the bludger or else QUIDITCH (that's numbawhang)!!!
func getWindowsUpdateSearcher() (*ole.IDispatch, error) {
	classID, err := oleutil.ClassIDFrom("Microsoft.Update.Session")
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != SFALSE {
			return &ole.IDispatch{}, err
		}
	}
	session, err := ole.CreateInstance(classID, ole.IID_IUnknown)
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != SFALSE {
			return &ole.IDispatch{}, err
		}
	}
	dispatch := session.MustQueryInterface(ole.IID_IDispatch)
	updateSearcherVar, err := dispatch.CallMethod("CreateUpdateSearcher")
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != SFALSE {
			return &ole.IDispatch{}, err
		}
	}
	updateSearcher := updateSearcherVar.ToIDispatch()
	return updateSearcher, nil

}

func searchUpdates(query string) (*ole.IEnumVARIANT, error) {
	updateSearcher, err := getWindowsUpdateSearcher()
	defer updateSearcher.Release()
	if err != nil {
		return nil, err
	}
	res, err := updateSearcher.CallMethod("Search", query)
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != SFALSE {
			return nil, err
		}
	}
	Updates, err := res.ToIDispatch().GetProperty("Updates")
	if err != nil {
		return nil, err
	}
	return getEnumFromDispatch(Updates.ToIDispatch())
}

func getEnumFromDispatch(dis *ole.IDispatch) (*ole.IEnumVARIANT, error) {
	listy, err := dis.GetProperty("_NewEnum")
	if err != nil {
		return nil, err
	}
	enum, err := listy.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return nil, err
	}
	return enum, nil
}
