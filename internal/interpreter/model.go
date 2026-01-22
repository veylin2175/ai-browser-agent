package interpreter

type PageState struct {
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Elements []Element `json:"elements"`
}

type Element struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
	Selector string `json:"selector"`
}
