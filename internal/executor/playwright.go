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
		"оплатить", "купить", "заказать", "подтвердить", "удалить", "удаление", "delete", "remove", "pay", "checkout", "оформить", "buy", "submit order",
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
		sel := els[a.Target].Selector
		return e.page.Locator(sel).Click()

	case core.ActionTypeText:
		if a.Target < 0 || a.Target >= len(els) {
			return fmt.Errorf("invalid target index: %d", a.Target)
		}

		sel := els[a.Target].Selector
		loc := e.page.Locator(sel)

		if err = loc.Click(); err != nil {
			return fmt.Errorf("не удалось сфокусироваться на элементе %d: %w", a.Target, err)
		}

		if err = loc.Fill(""); err != nil {
			log.Printf("Не удалось очистить поле: %v", err)
		}

		if err = loc.Fill(a.Text); err != nil {
			return fmt.Errorf("не удалось ввести текст: %w", err)
		}

		isSearchField := false
		nameLower := strings.ToLower(els[a.Target].Name)
		roleLower := strings.ToLower(els[a.Target].Role)

		if strings.Contains(nameLower, "поиск") ||
			strings.Contains(nameLower, "search") ||
			strings.Contains(roleLower, "searchbox") ||
			strings.Contains(roleLower, "combobox") && (strings.Contains(nameLower, "найти") || strings.Contains(nameLower, "search")) {
			isSearchField = true
		}

		if isSearchField {
			log.Println("Detected search-like input → pressing Enter to submit form")
			time.Sleep(600 * time.Millisecond) // чуть больше паузы — иногда помогает
			if err := e.page.Keyboard().Press("Enter"); err != nil {
				return fmt.Errorf("failed to press Enter: %w", err)
			}
			log.Println("Enter pressed successfully")
			time.Sleep(800 * time.Millisecond)
		}

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
