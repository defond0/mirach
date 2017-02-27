package parsers

import (
	"fmt"

	ole "github.com/go-ole/go-ole"
	oleutil "github.com/go-ole/go-ole/oleutil"
)

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
// https://github.com/StackExchange/wmi/blob/master/wmi.go#L54
const S_FALSE = 0x00000001

//grep returns exit status 1 when it gets no match, errors like that are fine
func GetWindowsKBs() (map[string][]KBArticle, []error) {
	err := coInit()
	defer ole.CoUninitialize()
	errors := []error{}
	out := make(map[string][]KBArticle)
	avail, err := getWindowsAvailableKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["available"] = avail
	avail_sec, err := getWindowsAvailableSecurityKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["available_security"] = avail_sec
	installed, err := getWindowsInstalledKBs()
	if err != nil {
		errors = append(errors, err)
	}
	out["installed"] = installed
	return out, errors
}

func getWindowsInstalledKBs() ([]KBArticle, error) {
	art := []KBArticle{}
	updates := searchUpdates("IsInstalled=1")
	defer updates.Release()
	var kbId string
	for update, length, err := updates.Next(1); length > 0; update, length, err = updates.Next(1) {
		if err != nil {
			fmt.Println(err)
		}
		defer update.Clear()
		update_dispatch := update.ToIDispatch()
		defer update_dispatch.Release()
		sev, err := update_dispatch.GetProperty("MsrcSeverity")
		kbs, err := update_dispatch.GetProperty("KBArticleIDs")
		if err != nil {
			fmt.Println(err)
		}
		kbIds := getEnumFromDispatch(kbs.ToIDispatch())
		for kb, length, _ := kbIds.Next(1); length > 0; kb, length, err = kbIds.Next(1) {
			defer kb.Clear()
			newKbId := fmt.Sprintf("KB%s", kb.Value())
			if newKbId != kbId {
				kbId = newKbId
				security := sev.Value() == "Critical"
				art = append(art, KBArticle{
					Name:     kbId,
					Security: security,
				},
				)
			}

		}
	}
	return art, nil
}

func getWindowsAvailableKBs() ([]KBArticle, error) {
	art := []KBArticle{}
	updates := searchUpdates("IsInstalled=0")
	defer updates.Release()
	var kbId string
	for update, length, err := updates.Next(1); length > 0; update, length, err = updates.Next(1) {
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		defer update.Clear()
		update_dispatch := update.ToIDispatch()
		defer update_dispatch.Release()
		sev, err := update_dispatch.GetProperty("MsrcSeverity")
		kbs, err := update_dispatch.GetProperty("KBArticleIDs")
		if err != nil {
			fmt.Println("prop error")
			fmt.Println(err)
		}
		kbIds := getEnumFromDispatch(kbs.ToIDispatch())
		for kb, length, _ := kbIds.Next(1); length > 0; kb, length, err = kbIds.Next(1) {
			newKbId := fmt.Sprintf("KB%s", kb.Value())
			if newKbId != kbId {
				kbId = newKbId
				security := sev.Value() == "Critical"
				art = append(art, KBArticle{
					Name:     kbId,
					Security: security,
				},
				)
			}

		}

	}
	return art, nil
}

//https://msdn.microsoft.com/en-us/library/windows/desktop/aa386906(v=vs.85).aspx
func getWindowsAvailableSecurityKBs() ([]KBArticle, error) {
	art, err := getWindowsAvailableKBs()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	artSec := []KBArticle{}
	for _, v := range art {
		if v.Security {
			artSec = append(artSec, v)
		}
	}
	return artSec, nil
}

func coInit() error {
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

//seek the bludger or else QUIDITCH (that's numbawhang)!!!
func getWindowsUpdateSearcher() (*ole.IDispatch, error) {
	classId, err := oleutil.ClassIDFrom("Microsoft.Update.Session")
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return &ole.IDispatch{}, err
		}
	}
	session, err := ole.CreateInstance(classId, ole.IID_IUnknown)
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return &ole.IDispatch{}, err
		}
	}
	dispatch := session.MustQueryInterface(ole.IID_IDispatch)
	updateSearcherVar, err := dispatch.CallMethod("CreateUpdateSearcher")
	if err != nil {
		fmt.Println(err)
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return &ole.IDispatch{}, err
		}
	}
	updateSearcher := updateSearcherVar.ToIDispatch()
	return updateSearcher, nil

}

func searchUpdates(query string) *ole.IEnumVARIANT {
	updateSearcher, err := getWindowsUpdateSearcher()
	defer updateSearcher.Release()
	if err != nil {
		fmt.Println("err getting update searcher")
		panic(err)
	}
	res, err := updateSearcher.CallMethod("Search", query)
	if err != nil {
		fmt.Println("UpdateSearcher.Search err")
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			panic(err)
		}
	}
	Updates, err := res.ToIDispatch().GetProperty("Updates")
	if err != nil {
		panic(err)
	}
	return getEnumFromDispatch(Updates.ToIDispatch())
}

func getEnumFromDispatch(dis *ole.IDispatch) *ole.IEnumVARIANT {
	listy, err := dis.GetProperty("_NewEnum")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	enum, err := listy.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return enum
}
