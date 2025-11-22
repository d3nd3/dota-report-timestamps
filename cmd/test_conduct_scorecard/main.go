package main

import (
	"bufio"
	"encoding/json"
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

	go func() {
		if err := client.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	}()

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
					fmt.Print("\nEnter Steam Guard Code: ")
					code, _ := reader.ReadString('\n')
					code = strings.TrimSpace(code)
					if code != "" {
						client.SubmitCode(code)
					}
				}

				if status == dota2gc.StatusGCReady {
					fmt.Println("\nGC is READY! Testing conduct scorecard API...")
					runConductScorecardTest(client)
					return
				}
			}
		}
	}
}

func runConductScorecardTest(client *dota2gc.Client) {
	fmt.Println("\nFetching conduct scorecard...")
	scorecard, err := client.GetPlayerConductScorecard()
	if err != nil {
		log.Fatalf("Error fetching conduct scorecard: %v", err)
	}

	fmt.Println("\n=== Conduct Scorecard ===")
	
	if scorecard.AccountId != nil {
		fmt.Printf("Account ID: %d\n", *scorecard.AccountId)
	}
	if scorecard.MatchId != nil {
		fmt.Printf("Match ID: %d\n", *scorecard.MatchId)
	}
	if scorecard.SeqNum != nil {
		fmt.Printf("Sequence Number: %d\n", *scorecard.SeqNum)
	}
	if scorecard.Reasons != nil {
		fmt.Printf("Reasons: %d\n", *scorecard.Reasons)
	}
	if scorecard.MatchesInReport != nil {
		fmt.Printf("Matches in Report: %d\n", *scorecard.MatchesInReport)
	}
	if scorecard.MatchesClean != nil {
		fmt.Printf("Matches Clean: %d\n", *scorecard.MatchesClean)
	}
	if scorecard.MatchesReported != nil {
		fmt.Printf("Matches Reported: %d\n", *scorecard.MatchesReported)
	}
	if scorecard.MatchesAbandoned != nil {
		fmt.Printf("Matches Abandoned: %d\n", *scorecard.MatchesAbandoned)
	}
	if scorecard.ReportsCount != nil {
		fmt.Printf("Reports Count: %d\n", *scorecard.ReportsCount)
	}
	if scorecard.ReportsParties != nil {
		fmt.Printf("Reports Parties: %d\n", *scorecard.ReportsParties)
	}
	if scorecard.CommendCount != nil {
		fmt.Printf("Commend Count: %d\n", *scorecard.CommendCount)
	}
	if scorecard.Date != nil {
		fmt.Printf("Date: %d (Unix timestamp)\n", *scorecard.Date)
		if *scorecard.Date > 0 {
			t := time.Unix(int64(*scorecard.Date), 0)
			fmt.Printf("Date (formatted): %s\n", t.Format(time.RFC3339))
		}
	}
	if scorecard.RawBehaviorScore != nil {
		fmt.Printf("Raw Behavior Score: %d\n", *scorecard.RawBehaviorScore)
	}
	if scorecard.OldRawBehaviorScore != nil {
		fmt.Printf("Old Raw Behavior Score: %d\n", *scorecard.OldRawBehaviorScore)
	}
	if scorecard.CommsReports != nil {
		fmt.Printf("Comms Reports: %d\n", *scorecard.CommsReports)
	}
	if scorecard.CommsParties != nil {
		fmt.Printf("Comms Parties: %d\n", *scorecard.CommsParties)
	}
	if scorecard.BehaviorRating != nil {
		ratingStr := "Unknown"
		switch *scorecard.BehaviorRating {
		case 0:
			ratingStr = "Good"
		case 1:
			ratingStr = "Warning"
		case 2:
			ratingStr = "Bad"
		}
		fmt.Printf("Behavior Rating: %d (%s)\n", *scorecard.BehaviorRating, ratingStr)
	}

	fmt.Println("\n=== Full JSON Response ===")
	jsonData, err := json.MarshalIndent(scorecard, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Println(string(jsonData))
	}

	if scorecard.MatchesInReport != nil {
		fmt.Printf("\n=== Analysis ===\n")
		fmt.Printf("This scorecard covers the last %d matches.\n", *scorecard.MatchesInReport)
		if *scorecard.MatchesInReport == 15 {
			fmt.Println("This appears to be a standard 15-game conduct report period.")
		}
		if scorecard.MatchId != nil {
			fmt.Printf("The report period ends at match ID: %d\n", *scorecard.MatchId)
		}
	}
}

