package parsers

import (
	"fmt"
	"os/exec"

	ole "github.com/go-ole/go-ole"
	oleutil "github.com/go-ole/go-ole/oleutil"
)

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
// https://github.com/StackExchange/wmi/blob/master/wmi.go#L54
const S_FALSE = 0x00000001

//grep returns exit status 1 when it gets no match, errors like that are fine
func GetWindowsKBs() (map[string][]KBArticle, []error) {
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
	wmicList := command("wmic path win32_quickfixengineering get HotFixId, Description")
	stdout, stderr, err := pipeline(wmicList)
	if err != nil {
		fmt.Println(string(stderr))
		return nil, err
	}
	return parseArticlesFromBytes(stdout)
}

func getWindowsAvailableKBs() ([]KBArticle, error) {
	art := []KBArticle{}
	fmt.Println("about to coinit")
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	defer ole.CoUninitialize()
	if err != nil {
		fmt.Println("coinit err")
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, err
		}
	}
	fmt.Println("conit seems to have worked")
	fmt.Println("Microsoft.Update.Session lets try and do that")
	classId, err := oleutil.ClassIDFrom("Microsoft.Update.Session")
	if err != nil {
		fmt.Println("classid err")
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, err
		}
	}
	session, err := ole.CreateInstance(classId, ole.IID_IUnknown)
	if err != nil {
		fmt.Println("Microsoft.Update.Session err")
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, err
		}
	}
	defer session.Release()
	dispatch := session.MustQueryInterface(ole.IID_IDispatch)
	defer dispatch.Release()
	fmt.Println("Microsoft.Update.Session seems to have worked with: ")
	fmt.Println(classId)
	fmt.Println(session)
	fmt.Println(dispatch)
	fmt.Println("updateSession.CreateUpdateSearcher lets try and do that")
	updateSearcher, err := dispatch.CallMethod("CreateUpdateSearcher")
	if err != nil {
		fmt.Println("CreateUpdateSearcher err")
		fmt.Println(err)
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, err
		}
	}
	defer updateSearcher.Clear()
	fmt.Println("updateSession.CreateUpdateSearcher seemed to work with:")
	fmt.Println(updateSearcher)
	fmt.Println("UpdateSearcher.Search IsInstalled=0 lets try and do that")
	res, err := updateSearcher.ToIDispatch().CallMethod("Search", "IsInstalled=0")
	defer res.Clear()
	if err != nil {
		fmt.Println("UpdateSearcher.Seartch IsInstalled=0 err")
		panic(err)
	}
	fmt.Println(res)
	Updates, err := res.ToIDispatch().GetProperty("Updates")
	defer Updates.Clear()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	listy, err := Updates.ToIDispatch().GetProperty("_NewEnum")
	defer listy.Clear()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	enum, err := listy.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	defer enum.Release()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	for tmp, length, err := enum.Next(1); length > 0; tmp, length, err = enum.Next(1) {
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		defer tmp.Clear()
		tmp_dispatch := tmp.ToIDispatch()
		defer tmp_dispatch.Release()
		type_, err := tmp_dispatch.GetProperty("Title")
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		ids, err := tmp.ToIDispatch().GetProperty("KBArticleIDs")
		kbs, err := ids.ToIDispatch().GetProperty("_NewEnum")
		defer kbs.Clear()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		kbs_enum, err := kbs.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
		defer kbs_enum.Release()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		for kb, length, err := kbs_enum.Next(1); length > 0; tmp, length, err = kbs_enum.Next(1) {
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			fmt.Println(kb.Value())
		}
		fmt.Println("Update")
		fmt.Println(type_.Value())
	}
	return art, nil
}

func getWindowsAvailableSecurityKBs() ([]KBArticle, error) {
	aptget := command("yum list updates -q --security")
	grep := exec.Command("grep", "-v", "Updated KBs")
	awk := exec.Command("awk", "{{ print $1 , $2 }}")
	stdout, stderr, err := pipeline(aptget, grep, awk)
	if err != nil {
		fmt.Println(string(stderr))
		return nil, err
	}
	return parseArticlesFromBytes(stdout)
}
