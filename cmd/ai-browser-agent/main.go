package main

import (
	"ai-browser-agent/internal/browser"
	"ai-browser-agent/internal/config"
	"ai-browser-agent/internal/interpreter"
	"fmt"
	"log"
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

	elements, err := interp.Snapshot()
	if err != nil {
		log.Fatal(err)
	}

	if len(elements) == 0 {
		log.Println("No interactive elements found")
	} else {
		fmt.Printf("Found %d elements\n", len(elements))
		fmt.Println(elements[0])
		_ = interp.Click(elements[0])
	}

	select {}
}
