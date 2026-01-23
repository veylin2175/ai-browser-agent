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
	"log"
	"os"
	"strings"
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
			fmt.Println("Ошибка при выполнении. Сейчас попробую снова:", err)
		}

		previousURL := br.Page.URL()
		previousTitle, _ := br.Page.Title()

		var observation string

		if err != nil {
			observation = fmt.Sprintf("ОШИБКА: %v. Действие не выполнено.", err)
		} else {
			newURL := br.Page.URL()
			newTitle, _ := br.Page.Title()

			changeHint := ""
			if newURL != previousURL {
				changeHint += fmt.Sprintf("URL изменился → %s\n", newURL)
			}
			if newTitle != previousTitle {
				changeHint += fmt.Sprintf("Заголовок изменился → %q\n", newTitle)
			}

			visibleText, errTxt := br.Page.Locator("body").InnerText()
			if errTxt == nil {
				visibleText = strings.ReplaceAll(visibleText, "\n", " ")
				visibleText = strings.TrimSpace(visibleText)
				if len(visibleText) > 400 {
					visibleText = visibleText[:350] + "... (обрезано)"
				}
				if len(visibleText) > 80 {
					changeHint += "Видимый текст страницы (начало): " + visibleText + "\n"
				}
			}

			if strings.Contains(newURL, "search") || strings.Contains(newURL, "yandex.ru") {
				results, _ := br.Page.Locator("h2, .organic__url, .text-container, .serp-item").AllInnerTexts()
				if len(results) > 0 && len(results[0]) > 10 {
					changeHint += "Обнаружены результаты поиска (первые заголовки): " +
						strings.Join(results[:min(3, len(results))], " | ") + "\n"
				}
			}

			if changeHint == "" {
				changeHint = "Состояние страницы почти не изменилось"
			}

			observation = fmt.Sprintf(
				"Действие выполнено успешно.\n%s\nТекущий URL: %s\nЗаголовок: %q",
				changeHint, newURL, newTitle,
			)

			previousURL = newURL
			previousTitle = newTitle
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
