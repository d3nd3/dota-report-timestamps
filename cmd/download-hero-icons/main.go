package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var heroIdMapping = map[string]string{
	"zeus": "zuus",
}

var heroesWithoutIcon = map[string]bool{
	"primal_beast": true,
	"ringmaster":   true,
	"marci":        true,
	"muerta":       true,
	"dawnbreaker":  true,
}

func main() {
	heroes := []string{
		"antimage", "axe", "bane", "bloodseeker", "crystal_maiden", "drow_ranger",
		"earthshaker", "juggernaut", "mirana", "morphling", "nevermore", "phantom_lancer",
		"puck", "pudge", "razor", "sand_king", "storm_spirit", "sven", "tiny",
		"vengefulspirit", "windrunner", "zeus", "kunkka", "kez", "lina", "lion",
		"shadow_shaman", "slardar", "tidehunter", "witch_doctor", "lich", "riki",
		"enigma", "tinker", "sniper", "necrolyte", "warlock", "beastmaster", "queenofpain",
		"venomancer", "faceless_void", "skeleton_king", "death_prophet", "phantom_assassin",
		"pugna", "templar_assassin", "viper", "luna", "dragon_knight", "dazzle",
		"rattletrap", "leshrac", "furion", "life_stealer", "dark_seer", "clinkz",
		"omniknight", "enchantress", "huskar", "night_stalker", "broodmother", "bounty_hunter",
		"weaver", "jakiro", "batrider", "chen", "spectre", "ancient_apparition", "doom_bringer",
		"ursa", "spirit_breaker", "gyrocopter", "alchemist", "invoker", "silencer",
		"obsidian_destroyer", "lycan", "brewmaster", "shadow_demon", "lone_druid", "chaos_knight",
		"meepo", "treant", "ogre_magi", "undying", "rubick", "disruptor", "nyx_assassin",
		"naga_siren", "keeper_of_the_light", "wisp", "visage", "slark", "medusa", "troll_warlord",
		"centaur", "magnataur", "shredder", "bristleback", "tusk", "skywrath_mage", "abaddon",
		"elder_titan", "legion_commander", "techies", "ember_spirit", "earth_spirit", "abyssal_underlord",
		"terrorblade", "phoenix", "oracle", "winter_wyvern", "arc_warden", "monkey_king",
		"dark_willow", "pangolier", "grimstroke", "mars", "snapfire", "void_spirit", "hoodwink",
		"dawnbreaker", "marci", "primal_beast", "muerta", "ringmaster",
	}

	assetsDir := filepath.Join("..", "..", "assets", "portraits")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	successCount := 0
	failCount := 0

	for _, heroId := range heroes {
		actualHeroId := heroId
		if mapped, ok := heroIdMapping[heroId]; ok {
			actualHeroId = mapped
		}

		var iconUrl string
		if heroId == "kez" || heroesWithoutIcon[heroId] {
			iconUrl = fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", actualHeroId)
		} else {
			iconUrl = fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_icon.png", actualHeroId)
		}

		filePath := filepath.Join(assetsDir, fmt.Sprintf("%s.png", heroId))

		fmt.Printf("Downloading %s... ", heroId)
		resp, err := client.Get(iconUrl)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failCount++
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if !heroesWithoutIcon[heroId] && heroId != "kez" {
				fallbackUrl := fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", actualHeroId)
				resp.Body.Close()
				resp, err = client.Get(fallbackUrl)
				if err != nil {
					fmt.Printf("FAILED: %v\n", err)
					failCount++
					continue
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					fmt.Printf("FAILED: HTTP %d\n", resp.StatusCode)
					failCount++
					continue
				}
			} else {
				fmt.Printf("FAILED: HTTP %d\n", resp.StatusCode)
				failCount++
				continue
			}
		}

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failCount++
			continue
		}

		_, err = io.Copy(file, resp.Body)
		file.Close()
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failCount++
			os.Remove(filePath)
			continue
		}

		fmt.Printf("OK\n")
		successCount++
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\nDownload complete: %d succeeded, %d failed\n", successCount, failCount)
}

