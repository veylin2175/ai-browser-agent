package interpreter

type Element struct {
	Index      int    `json:"index"`
	Selector   string `json:"selector"`
	Role       string `json:"role"`
	Name       string `json:"name"`
	Disabled   bool   `json:"disabled"`
	Visible    bool   `json:"visible"`
	IsHidden   bool   `json:"isHidden"`
	InViewport bool   `json:"inViewport"`
}
