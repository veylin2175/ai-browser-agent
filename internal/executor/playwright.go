package executor

import (
	"ai-browser-agent/internal/core"
	"fmt"
	"log"
	"strings"
	"time"

	"ai-browser-agent/internal/interpreter"

	"github.com/playwright-community/playwright-go"
)

type PlaywrightExecutor struct {
	page playwright.Page
	i    *interpreter.Interpreter
}

func New(page playwright.Page, i *interpreter.Interpreter) *PlaywrightExecutor {
	return &PlaywrightExecutor{page: page, i: i}
}

func (e *PlaywrightExecutor) Execute(a *core.Action) error {
	els, err := e.i.Snapshot()
	if err != nil {
		return err
	}

	destructiveKeywords := []string{
		"оплатить", "купить", "заказать", "подтвердить", "удалить", "удаление",
		"delete", "remove", "pay", "checkout", "оформить", "buy", "submit order",
	}

	isDestructive := false
	destructiveReason := ""

	if a.Type == core.ActionClick || a.Type == core.ActionTypeText || a.Type == core.ActionPressKey {
		if a.Target >= 0 && a.Target < len(els) {
			el := els[a.Target]
			nameLower := strings.ToLower(el.Name)
			roleLower := strings.ToLower(el.Role)
			selectorLower := strings.ToLower(el.Selector)

			for _, kw := range destructiveKeywords {
				if strings.Contains(nameLower, kw) ||
					strings.Contains(roleLower, kw) ||
					strings.Contains(selectorLower, kw) {
					isDestructive = true
					destructiveReason = fmt.Sprintf("Элемент содержит подозрительное слово: %q (name=%q, role=%q)", kw, el.Name, el.Role)
					break
				}
			}
		}
	}

	if isDestructive {
		fmt.Printf("\n⚠️ ВНИМАНИЕ: потенциально деструктивное действие!\n")
		fmt.Printf("Действие: %s\n", a.String())
		fmt.Printf("Причина: %s\n", destructiveReason)
		fmt.Print("Подтвердить выполнение? (y/n): ")

		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		if input != "y" && input != "yes" {
			return fmt.Errorf("действие отменено пользователем")
		}

		log.Println("Действие подтверждено пользователем")
	}

	switch a.Type {
	case core.ActionClick:
		if a.Target < 0 || a.Target >= len(els) {
			return fmt.Errorf("invalid target index: %d (elements: %d)", a.Target, len(els))
		}

		el := els[a.Target]
		sel := el.Selector

		log.Printf("Попытка клика по элементу %d: selector=%s, name=%q, role=%s, inViewport=%v",
			a.Target, sel, el.Name, el.Role, el.InViewport)

		loc := e.page.Locator(sel).First()

		if err = loc.ScrollIntoViewIfNeeded(); err != nil {
			log.Printf("Предупреждение: не удалось проскроллить к элементу: %v", err)
		}

		time.Sleep(300 * time.Millisecond)

		if err = loc.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("элемент %d (%s) не стал видимым за 10с: %w", a.Target, sel, err)
		}

		if err = loc.Click(playwright.LocatorClickOptions{
			Timeout: playwright.Float(10000),
			Force:   playwright.Bool(false),
		}); err != nil {
			log.Printf("Обычный клик не сработал, пробуем force click: %v", err)
			if err = loc.Click(playwright.LocatorClickOptions{
				Timeout: playwright.Float(10000),
				Force:   playwright.Bool(true),
			}); err != nil {
				return fmt.Errorf("не удалось кликнуть даже с force: %w", err)
			}
		}

		time.Sleep(500 * time.Millisecond)
		return nil

	case core.ActionTypeText:
		if a.Target < 0 || a.Target >= len(els) {
			return fmt.Errorf("invalid target index: %d", a.Target)
		}

		el := els[a.Target]
		sel := el.Selector
		loc := e.page.Locator(sel).First()

		if err = loc.ScrollIntoViewIfNeeded(); err != nil {
			log.Printf("Предупреждение: не удалось проскроллить к полю ввода: %v", err)
		}

		time.Sleep(300 * time.Millisecond)

		if err = loc.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(10000),
		}); err != nil {
			return fmt.Errorf("поле ввода %d не стало видимым: %w", a.Target, err)
		}

		if err = loc.Click(); err != nil {
			return fmt.Errorf("не удалось сфокусироваться на элементе %d: %w", a.Target, err)
		}

		if err = loc.Fill(""); err != nil {
			log.Printf("Не удалось очистить поле: %v", err)
		}

		if err = loc.Fill(a.Text); err != nil {
			return fmt.Errorf("не удалось ввести текст: %w", err)
		}

		time.Sleep(300 * time.Millisecond)
		return nil

	case core.ActionNavigate:
		_, err = e.page.Goto(a.URL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(15000),
		})
		if err != nil {
			return err
		}
		time.Sleep(1500 * time.Millisecond)

	case core.ActionPressKey:
		if a.Key == "" {
			return fmt.Errorf("press_key требует поле 'key'")
		}
		return e.page.Keyboard().Press(a.Key)

	case core.ActionDone:
		return nil

	default:
		return fmt.Errorf("неизвестный тип действия: %s", a.Type)
	}

	return nil
}
