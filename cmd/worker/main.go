package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/menyasosali/mts/config"
	"github.com/menyasosali/mts/internal/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := &config.WorkerConfig{}

	cfgPath := "./config/config.yaml"

	err := cleanenv.ReadConfig(cfgPath, cfg)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	fmt.Println("Worker Config:")
	fmt.Printf("%+v\n", cfg)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(cfgPath)
	if err != nil {
		log.Fatalf("Failed to add config file to watcher: %v", err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					// Файл конфигурации был изменен
					fmt.Println("Config file changed. Reloading...")

					// Перечитываем конфигурацию
					err := cleanenv.ReadConfig(cfgPath, cfg)
					if err != nil {
						log.Printf("Failed to reload config file: %v", err)
					} else {
						fmt.Println("Config reloaded:")
						fmt.Printf("%+v\n", cfg)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	worker.Run(cfg)

	<-exit
	fmt.Println("Worker shutdown")
}
