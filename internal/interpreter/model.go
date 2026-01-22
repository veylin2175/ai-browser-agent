package interpreter

type PageState struct {
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Elements []Element `json:"elements"`
}

type Element struct {
	Index    int    `json:"index"`
	Selector string `json:"selector"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
	Visible  bool   `json:"visible"`
}
