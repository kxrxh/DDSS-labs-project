package main

import (
	"fmt"

	"github.com/kxrxh/ddss/internal/db"
)

func main() {
	cfg := db.Config{
		DgraphAddr: "localhost:9080",
		MongoURI:   "mongodb://localhost:27017",
		TiKVAddr:   []string{"localhost:2379", "localhost:2380", "localhost:2381"},
	}

	client, err := db.NewClient(cfg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer client.Close()

	// Add a citizen with relations
	_, err = client.AddCitizen("citizen001", "Alice", []string{"citizen002"})
	if err != nil {
		fmt.Println("Error adding citizen:", err)
	}

	// Add citizen profile
	err = client.AddCitizenProfile("citizen001", "Alice", 30, 100)
	if err != nil {
		fmt.Println("Error adding profile:", err)
	}

	// Blacklist someone
	err = client.AddToBlacklist("citizen003", "Fraud detected")
	if err != nil {
		fmt.Println("Error blacklisting:", err)
	}

	// Check blacklist
	isBlacklisted, reason, err := client.IsBlacklisted("citizen003")
	if err != nil {
		fmt.Println("Error checking blacklist:", err)
	} else {
		fmt.Printf("Blacklisted: %v, Reason: %s\n", isBlacklisted, reason)
	}

	select {}
}
