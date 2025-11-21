package main

import (
	"fmt"
	"net/http"
	"time"
)

var allHeroes = []struct {
	id   string
	name string
}{
	{"antimage", "Anti-Mage"},
	{"axe", "Axe"},
	{"bane", "Bane"},
	{"bloodseeker", "Bloodseeker"},
	{"crystal_maiden", "Crystal Maiden"},
	{"drow_ranger", "Drow Ranger"},
	{"earthshaker", "Earthshaker"},
	{"juggernaut", "Juggernaut"},
	{"mirana", "Mirana"},
	{"morphling", "Morphling"},
	{"nevermore", "Shadow Fiend"},
	{"phantom_lancer", "Phantom Lancer"},
	{"puck", "Puck"},
	{"pudge", "Pudge"},
	{"razor", "Razor"},
	{"sand_king", "Sand King"},
	{"storm_spirit", "Storm Spirit"},
	{"sven", "Sven"},
	{"tiny", "Tiny"},
	{"vengefulspirit", "Vengeful Spirit"},
	{"windrunner", "Windrunner"},
	{"zeus", "Zeus"},
	{"kunkka", "Kunkka"},
	{"kez", "Kez"},
	{"lina", "Lina"},
	{"lion", "Lion"},
	{"shadow_shaman", "Shadow Shaman"},
	{"slardar", "Slardar"},
	{"tidehunter", "Tidehunter"},
	{"witch_doctor", "Witch Doctor"},
	{"lich", "Lich"},
	{"riki", "Riki"},
	{"enigma", "Enigma"},
	{"tinker", "Tinker"},
	{"sniper", "Sniper"},
	{"necrolyte", "Necrophos"},
	{"warlock", "Warlock"},
	{"beastmaster", "Beastmaster"},
	{"queenofpain", "Queen of Pain"},
	{"venomancer", "Venomancer"},
	{"faceless_void", "Faceless Void"},
	{"skeleton_king", "Wraith King"},
	{"death_prophet", "Death Prophet"},
	{"phantom_assassin", "Phantom Assassin"},
	{"pugna", "Pugna"},
	{"templar_assassin", "Templar Assassin"},
	{"viper", "Viper"},
	{"luna", "Luna"},
	{"dragon_knight", "Dragon Knight"},
	{"dazzle", "Dazzle"},
	{"rattletrap", "Clockwerk"},
	{"leshrac", "Leshrac"},
	{"furion", "Nature's Prophet"},
	{"life_stealer", "Lifestealer"},
	{"dark_seer", "Dark Seer"},
	{"clinkz", "Clinkz"},
	{"omniknight", "Omniknight"},
	{"enchantress", "Enchantress"},
	{"huskar", "Huskar"},
	{"night_stalker", "Night Stalker"},
	{"broodmother", "Broodmother"},
	{"bounty_hunter", "Bounty Hunter"},
	{"weaver", "Weaver"},
	{"jakiro", "Jakiro"},
	{"batrider", "Batrider"},
	{"chen", "Chen"},
	{"spectre", "Spectre"},
	{"ancient_apparition", "Ancient Apparition"},
	{"doom_bringer", "Doom"},
	{"ursa", "Ursa"},
	{"spirit_breaker", "Spirit Breaker"},
	{"gyrocopter", "Gyrocopter"},
	{"alchemist", "Alchemist"},
	{"invoker", "Invoker"},
	{"silencer", "Silencer"},
	{"obsidian_destroyer", "Outworld Devourer"},
	{"lycan", "Lycan"},
	{"brewmaster", "Brewmaster"},
	{"shadow_demon", "Shadow Demon"},
	{"lone_druid", "Lone Druid"},
	{"chaos_knight", "Chaos Knight"},
	{"meepo", "Meepo"},
	{"treant", "Treant Protector"},
	{"ogre_magi", "Ogre Magi"},
	{"undying", "Undying"},
	{"rubick", "Rubick"},
	{"disruptor", "Disruptor"},
	{"nyx_assassin", "Nyx Assassin"},
	{"naga_siren", "Naga Siren"},
	{"keeper_of_the_light", "Keeper of the Light"},
	{"wisp", "Io"},
	{"visage", "Visage"},
	{"slark", "Slark"},
	{"medusa", "Medusa"},
	{"troll_warlord", "Troll Warlord"},
	{"centaur", "Centaur Warrunner"},
	{"magnataur", "Magnus"},
	{"shredder", "Timbersaw"},
	{"bristleback", "Bristleback"},
	{"tusk", "Tusk"},
	{"skywrath_mage", "Skywrath Mage"},
	{"abaddon", "Abaddon"},
	{"elder_titan", "Elder Titan"},
	{"legion_commander", "Legion Commander"},
	{"techies", "Techies"},
	{"ember_spirit", "Ember Spirit"},
	{"earth_spirit", "Earth Spirit"},
	{"abyssal_underlord", "Underlord"},
	{"terrorblade", "Terrorblade"},
	{"phoenix", "Phoenix"},
	{"oracle", "Oracle"},
	{"winter_wyvern", "Winter Wyvern"},
	{"arc_warden", "Arc Warden"},
	{"monkey_king", "Monkey King"},
	{"dark_willow", "Dark Willow"},
	{"pangolier", "Pangolier"},
	{"grimstroke", "Grimstroke"},
	{"hoodwink", "Hoodwink"},
	{"void_spirit", "Void Spirit"},
	{"snapfire", "Snapfire"},
	{"mars", "Mars"},
	{"dawnbreaker", "Dawnbreaker"},
	{"marci", "Marci"},
	{"primal_beast", "Primal Beast"},
	{"muerta", "Muerta"},
	{"ringmaster", "Ringmaster"},
}

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	heroesWithoutIcon := map[string]bool{
		"primal_beast": true,
		"ringmaster":   true,
		"marci":        true,
	}

	successCount := 0
	failCount := 0
	var failedHeroes []string

	fmt.Println("Testing hero icon availability from Steam CDN...")
	fmt.Println("============================================================")

	for _, hero := range allHeroes {
		var iconUrl string
		var useFullImage bool

		if hero.id == "kez" || heroesWithoutIcon[hero.id] {
			iconUrl = fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", hero.id)
			useFullImage = true
		} else {
			iconUrl = fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_icon.png", hero.id)
			useFullImage = false
		}

		resp, err := client.Get(iconUrl)
		if err != nil {
			fmt.Printf("❌ %s (%s): Error - %v\n", hero.name, hero.id, err)
			failCount++
			failedHeroes = append(failedHeroes, fmt.Sprintf("%s (%s)", hero.name, hero.id))

			if !useFullImage && !heroesWithoutIcon[hero.id] && hero.id != "kez" {
				fullUrl := fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", hero.id)
				resp2, err2 := client.Get(fullUrl)
				if err2 == nil {
					resp2.Body.Close()
					if resp2.StatusCode == http.StatusOK {
						fmt.Printf("   → Full image available: %s\n", fullUrl)
					}
				}
			}
			continue
		}

		statusCode := resp.StatusCode
		resp.Body.Close()

		if statusCode == http.StatusOK {
			fmt.Printf("✅ %s (%s): OK\n", hero.name, hero.id)
			successCount++
		} else {
			fmt.Printf("❌ %s (%s): Status %d\n", hero.name, hero.id, statusCode)
			failCount++
			failedHeroes = append(failedHeroes, fmt.Sprintf("%s (%s)", hero.name, hero.id))

			if !useFullImage && !heroesWithoutIcon[hero.id] && hero.id != "kez" {
				fullUrl := fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", hero.id)
				resp2, err2 := client.Get(fullUrl)
				if err2 == nil {
					statusCode2 := resp2.StatusCode
					resp2.Body.Close()
					if statusCode2 == http.StatusOK {
						fmt.Printf("   → Full image available: %s\n", fullUrl)
					} else {
						fmt.Printf("   → Full image also failed: Status %d\n", statusCode2)
					}
				}
			}
		}
	}

	fmt.Println("============================================================")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  ✅ Success: %d\n", successCount)
	fmt.Printf("  ❌ Failed:  %d\n", failCount)

	if len(failedHeroes) > 0 {
		fmt.Printf("\nFailed heroes:\n")
		for _, hero := range failedHeroes {
			fmt.Printf("  - %s\n", hero)
		}
	}
}

