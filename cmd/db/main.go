package main

import (
	"auth/internal/config"
	"auth/internal/entity"
	"auth/internal/storage/postgres"
	"log"
)








func main() {

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.ConnectToDb(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entity.User{}); err != nil {
		log.Fatalf("failed to migrate")
	}

}