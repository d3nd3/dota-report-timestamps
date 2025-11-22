package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/dota2gc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	reader := bufio.NewReader(os.Stdin)

	username := os.Getenv("STEAM_USER")
	password := os.Getenv("STEAM_PASS")

	if username == "" {
		fmt.Print("Enter Steam Username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
	}

	if password == "" {
		fmt.Print("Enter Steam Password: ")
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)
	}

	fmt.Printf("Initializing Dota 2 GC Client for user: %s\n", username)
	client := dota2gc.NewClient(username, password)

	// Start connection in a goroutine so we can monitor status
	go func() {
		if err := client.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	}()

	// Main loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastStatus dota2gc.ConnectionStatus = -1

	for {
		select {
		case <-ticker.C:
			status := client.GetStatus()
			if status != lastStatus {
				fmt.Printf("Status changed to: %d\n", status)
				lastStatus = status

				if status == dota2gc.StatusNeedGuardCode {
					fmt.Print("\nCreate Steam Guard Code: ")
					code, _ := reader.ReadString('\n')
					code = strings.TrimSpace(code)
					if code != "" {
						client.SubmitCode(code)
					}
				}

				if status == dota2gc.StatusGCReady {
					fmt.Println("\nGC is READY! You can now fetch match history.")
					runMatchHistoryTest(client, reader)
					// After test, exit or loop? Let's loop so user can run again
				}
			}
		}
	}
}

func runMatchHistoryTest(client *dota2gc.Client, reader *bufio.Reader) {
	for {
		fmt.Print("\nEnter SteamID64 to fetch history (or 'q' to quit, 'p' to test pagination): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" {
			os.Exit(0)
		}

		if input == "p" {
			// Test pagination
			testPagination(client, reader)
			continue
		}

		if input == "" {
			continue
		}

		var steamID int64
		fmt.Sscanf(input, "%d", &steamID)

		if steamID == 0 {
			fmt.Println("Invalid SteamID")
			continue
		}

		fmt.Printf("Fetching history for %d...\n", steamID)
		matches, err := client.GetPlayerMatchHistory(steamID, 10, false)
		if err != nil {
			log.Printf("Error fetching history: %v", err)
			continue
		}

		fmt.Printf("Found %d matches:\n", len(matches))
		for i, m := range matches {
			fmt.Printf("%d: MatchID=%d, StartTime=%d, LobbyType=%d, GameMode=%d\n", i, m.ID, m.StartTime, m.LobbyType, m.GameMode)
		}
	}
}

func testPagination(client *dota2gc.Client, reader *bufio.Reader) {
	fmt.Print("Enter SteamID64 for pagination test: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	var steamID int64
	fmt.Sscanf(input, "%d", &steamID)

	if steamID == 0 {
		fmt.Println("Invalid SteamID")
		return
	}

	fmt.Println("Fetching first page (limit 5)...")
	matches1, err := client.GetPlayerMatchHistory(steamID, 5, false)
	if err != nil {
		log.Printf("Error fetching page 1: %v", err)
		return
	}
	
	if len(matches1) == 0 {
		fmt.Println("No matches found.")
		return
	}

	lastMatchID := matches1[len(matches1)-1].ID
	fmt.Printf("Page 1 returned %d matches. Last Match ID: %d\n", len(matches1), lastMatchID)
	for i, m := range matches1 {
		fmt.Printf("P1[%d]: %d\n", i, m.ID)
	}

	fmt.Printf("\nFetching second page (startAtMatchID=%d)...\n", lastMatchID)
	matches2, err := client.GetPlayerMatchHistoryPaginated(steamID, 5, false, uint64(lastMatchID))
	if err != nil {
		log.Printf("Error fetching page 2: %v", err)
		return
	}

	fmt.Printf("Page 2 returned %d matches:\n", len(matches2))
	for i, m := range matches2 {
		fmt.Printf("P2[%d]: %d\n", i, m.ID)
	}
	
	if len(matches2) > 0 {
		if matches2[0].ID == lastMatchID {
			fmt.Println("\nRESULT: Pagination is INCLUSIVE (first match of P2 == last match of P1)")
		} else {
			fmt.Println("\nRESULT: Pagination is EXCLUSIVE (first match of P2 != last match of P1)")
		}
	}
}

