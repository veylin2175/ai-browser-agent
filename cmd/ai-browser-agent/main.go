package main

import (
	"ai-browser-agent/internal/browser"
	"ai-browser-agent/internal/config"
	"ai-browser-agent/internal/interpreter"
	"log"
)

func main() {
	// Загружаем конфиг
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

	interp.PrintSnapshot()

	_, err = interp.Snapshot()
	if err != nil {
		log.Fatal(err)
	}

	select {}

}
