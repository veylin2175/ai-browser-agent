package interpreter

func (i *Interpreter) Click(element Element) error {
	return i.page.Locator(element.Selector).First().Click()
}

func (i *Interpreter) Type(element Element, text string) error {
	loc := i.page.Locator(element.Selector).First()
	if err := loc.Click(); err != nil {
		return err
	}
	return loc.Fill(text)
}

func (i *Interpreter) Navigate(url string) error {
	_, err := i.page.Goto(url)
	return err
}
