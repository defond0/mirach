package parsers

//LinuxPackage represent pertinent features of linux package
type LinuxPackage struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Security bool   `json:"security"`
}

//KBArticle represent pertinent features of linux package
type KBArticle struct {
	Name     string `json:"name"`
	Security bool   `json:"security"`
}
