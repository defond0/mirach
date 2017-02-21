package parsers

import (
	"fmt"
	"os/exec"
	"reflect"

	ole "github.com/go-ole/go-ole"
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
	err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	defer ole.CoUninitialize()
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, err
		}
	}
	session, err := ole.GetActiveObject("Microsoft.Update.Session")
	defer session.Release()
	fmt.Println("session")
	fmt.Println(session)
	fmt.Println(reflect.TypeOf(session))
	if err != nil {
		fmt.Println("err")
		oleCode := err.(*ole.OleError).Code()
		fmt.Println(err)
		fmt.Println(oleCode)
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, err
		}
	}
	update, err := session.QueryInterface(ole.IID_IDispatch)
	defer update.Release()
	fmt.Println("update")
	fmt.Println(update)
	fmt.Println(reflect.TypeOf(update))
	// temp, err := update.CallMethod("CreateupdateSearcher")
	// if err != nil {
	// 	fmt.Println("err")
	// 	oleCode := err.(*ole.OleError).Code()
	// 	fmt.Println(err)
	// 	fmt.Println(oleCode)
	// 	if oleCode != ole.S_OK && oleCode != S_FALSE {
	// 		return nil, err
	// 	}
	// }
	// search := temp.ToIDispatch()
	// defer search.Release()
	// fmt.Println("search")
	// fmt.Println(search)
	// fmt.Println(search.GetTypeInfoCount())
	// res, err := search.CallMethod("Search", "IsInstalled=0")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("res")
	// fmt.Println(res)
	// resdis := res.ToIDispatch()
	// fmt.Println(resdis.GetTypeInfo())
	// fmt.Println(resdis.GetTypeInfoCount())
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
