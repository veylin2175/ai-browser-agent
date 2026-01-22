package interpreter

func (i *Interpreter) Navigate(url string) error {
	_, err := i.page.Goto(url)
	return err
}
