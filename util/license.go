package util

import "fmt"

var mirachLicense = "Copyright 2017 ClearDATA"

// ShowLicenseInfo will output mirach's license text.
// mirach does not use any third party software that requires license
// inclusion for using the library, and the libraries used are not
// included in the distribution.
func ShowLicenseInfo() {
	fmt.Println(mirachLicense)
}
