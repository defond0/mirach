package parsers

//LinuxPackage represent pertinent features of linux package
type LinuxPackage struct {
	name     string
	Version  string `json:"version"`
	Security bool   `json:"security"`
}

//KBArticle represent pertinent features of linux package
type KBArticle struct {
	name     string
	Security bool `json:"security"`
}
