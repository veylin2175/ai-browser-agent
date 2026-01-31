package main

import (
	"ai-browser-agent/internal/agent"
	"ai-browser-agent/internal/browser"
	"ai-browser-agent/internal/config"
	"ai-browser-agent/internal/core"
	"ai-browser-agent/internal/executor"
	"ai-browser-agent/internal/interpreter"
	"ai-browser-agent/internal/llm"
	"bufio"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	cfg, err := config.Load("config/local.yml")
	if err != nil {
		log.Fatal(err)
	}

	br, err := browser.Launch(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer br.Close()

	br.Page.SetDefaultTimeout(10000)

	_, err = br.Page.Goto("https://example.com")
	if err != nil {
		log.Fatal(err)
	}

	interp := interpreter.New(br.Page)
	exec := executor.New(br.Page, interp)

	llmClient := llm.NewZai(cfg)

	ag := agent.New(llmClient, interp)

	fmt.Println("Введите цель для агента (нажмите Enter после ввода):")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	goal := strings.TrimSpace(scanner.Text())

	if goal == "" {
		log.Fatal("Цель не введена. Завершение.")
	}

	fmt.Println("Цель получена. Агент начинает работу...")

	for {
		action, err := ag.Step(goal)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("→ %s\n", action.String())

		if action.Type == core.ActionDone {
			break
		}

		if err = exec.Execute(action); err != nil {
			fmt.Printf("! Ошибка выполнения действия: %v\n", err)
		}

		if err == nil {
			_ = br.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State:   playwright.LoadStateDomcontentloaded,
				Timeout: playwright.Float(10000),
			})
			time.Sleep(1200 * time.Millisecond)
		}

		var observation string

		if err != nil {
			observation = fmt.Sprintf("ОШИБКА: %v", err)
		} else {
			currentURL := br.Page.URL()
			currentTitle, _ := br.Page.Title()

			visibleText, errTxt := br.Page.Locator("body").InnerText()
			if errTxt == nil {
				visibleText = strings.ReplaceAll(visibleText, "\n", " ")
				visibleText = strings.TrimSpace(visibleText)
				if len(visibleText) > 400 {
					visibleText = visibleText[:350] + "... (обрезано)"
				}
			} else {
				visibleText = "(не удалось получить текст)"
			}

			observation = fmt.Sprintf(
				"Действие выполнено.\nURL: %s\nЗаголовок: %q\nВидимый текст (начало): %s",
				currentURL, currentTitle, visibleText,
			)
		}

		ag.History = append(ag.History, fmt.Sprintf("%s → %s", action.String(), observation))

		if len(ag.History) > 10 {
			ag.History = ag.History[len(ag.History)-10:]
		}
	}

	fmt.Println("Нажмите Enter в терминале, чтобы закрыть браузер и завершить программу...")
	var input string
	fmt.Scanln(&input)
}
