package main

import (
	"github.com/zhangkui/go-home-inventory/internal/httpapi"
	"github.com/zhangkui/go-home-inventory/internal/inventory"
	"github.com/zhangkui/go-home-inventory/internal/location"
	"github.com/zhangkui/go-home-inventory/internal/reminder"
	"github.com/zhangkui/go-home-inventory/internal/repair"
	"github.com/zhangkui/go-home-inventory/internal/warranty"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	zoneName := os.Getenv("APP_TIMEZONE")
	if zoneName == "" {
		zoneName = "Asia/Shanghai"
	}
	zone, err := time.LoadLocation(zoneName)
	if err != nil {
		log.Fatalf("load timezone: %v", err)
	}
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	store := inventory.NewStore(time.Now)
	warranties := warranty.NewService(store, zone, time.Now)
	handler := httpapi.NewServer(store, location.NewService(store), warranties, repair.NewService(store, time.Now), reminder.NewService(store, warranties, zone, time.Now), zone)
	server := &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s", address)
	log.Fatal(server.ListenAndServe())
}
