package parsers

type LinuxPackage struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Security bool   `json:"security"`
}

type KBArticle struct {
	Name     string `json:"name"`
	Security bool   `json:"security"`
}
