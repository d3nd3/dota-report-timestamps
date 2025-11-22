package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/dotabuff/manta"
	"github.com/dotabuff/manta/dota"
	"github.com/golang/protobuf/proto"
	"github.com/klauspost/compress/snappy"
)

type PlayerResource struct {
	Slot     int    `json:"Slot"`
	SteamID  uint64 `json:"SteamID"`
	EntIndex uint32
	Team     int32  `json:"Team"` // 2 = radiant, 3 = dire
	Name     string `json:"Name"`
	Hero     string `json:"Hero"`
}

type Report struct {
	Time          string
	Team          string // "FRIENDLY" or "ENEMY"
	SteamID       uint64
	Slot          int
	Name          string
	Hero          string
	TargetSlot    int    `json:"TargetSlot"`    // The slot of the player who was reported
	TargetSteamID uint64 `json:"TargetSteamID"` // The SteamID of the player who was reported
	TargetName    string `json:"TargetName"`    // The name of the player who was reported
	TargetHero    string `json:"TargetHero"`    // The hero of the player who was reported
}

type ParseResult struct {
	MatchID      int64     `json:"MatchID"`
	TeamReports  int       `json:"TeamReports"`
	EnemyReports int       `json:"EnemyReports"`
	Reports      []*Report `json:"Reports"`
}

// reader performs read operations against a buffer
type reader struct {
	buf      []byte
	size     uint32
	pos      uint32
	bitVal   uint64 // value of the remaining bits in the current byte
	bitCount uint32 // number of remaining bits in the current byte
}

// nextByte reads the next byte from the buffer
func (r *reader) nextByte() byte {
	r.pos += 1
	if r.pos > r.size {
		panic("nextByte: insufficient buffer")
	}
	return r.buf[r.pos-1]
}

// readUBitVar reads a variable length uint32 with encoding in last to bits of 6 bit group
func (r *reader) readUBitVar() uint32 {
	ret := r.readBits(6)

	switch ret & 0x30 {
	case 16:
		ret = (ret & 15) | (r.readBits(4) << 4)
		break
	case 32:
		ret = (ret & 15) | (r.readBits(8) << 4)
		break
	case 48:
		ret = (ret & 15) | (r.readBits(28) << 4)
		break
	}

	return ret
}

// readBits returns the uint32 value for the given number of sequential bits
func (r *reader) readBits(n uint32) uint32 {
	for n > r.bitCount {
		r.bitVal |= uint64(r.nextByte()) << r.bitCount
		r.bitCount += 8
	}

	x := (r.bitVal & ((1 << n) - 1))
	r.bitVal >>= n
	r.bitCount -= n

	return uint32(x)
}

func ticksToMinutesAndSeconds(begin_tick int, pausedTicks int, ticks int) (int, int) {
	if begin_tick == 0 {
		ticks = 0
	} else {
		ticks = ticks - begin_tick - pausedTicks
	}
	totalSeconds := ticks / 30
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return minutes, seconds
}

func formatDuration(d time.Duration) string {
	return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC).Add(d).Truncate(time.Second).Format("15:04:05.999999999")
}

func isReportButton(x int, y int, aspect float32) int {
	// width of their scoreboard in 1920 pixels, based on thier custom resolution,aspectRatio.
	score_width := int(math.Round(float64((1.77777777778 * (920.0 / 1920)) / float64(aspect) * 1920)))
	score_width_no_tips := int(math.Round(float64((1.77777777778 * (820.0 / 1920)) / float64(aspect) * 1920)))

	lower_x_bound := 865.0
	upper_x_bound := 893.0

	lower_x_bound_r := lower_x_bound / 920
	upper_x_bound_r := upper_x_bound / 920

	lower_x_bound_no_tips_r := (lower_x_bound - 100) / 820
	upper_x_bound_no_tips_r := (upper_x_bound - 100) / 820

	if (x >= int(math.Floor(float64(lower_x_bound_no_tips_r*float64(score_width_no_tips)))) && x <= int(math.Ceil(float64(upper_x_bound_no_tips_r*float64(score_width_no_tips))))) || (x >= int(math.Floor(float64(lower_x_bound_r*float64(score_width)))) && x <= int(math.Ceil(float64(upper_x_bound_r*float64(score_width))))) /*|| ( x >= 658 && x <= 679 )*/ {
		if y >= 106 && y <= 134 { //120+70x
			return 0
		} else if y >= 176 && y <= 204 {
			return 1
		} else if y >= 246 && y <= 274 {
			return 2
		} else if y >= 316 && y <= 344 {
			return 3
		} else if y >= 386 && y <= 414 {
			return 4
		} else if y >= 486 && y <= 514 {
			return 5
		} else if y >= 556 && y <= 584 {
			return 6
		} else if y >= 626 && y <= 654 {
			return 7
		} else if y >= 696 && y <= 724 {
			return 8
		} else if y >= 766 && y <= 794 {
			return 9
		}
	}
	return -1
}

func ExtractPlayerInfo(matchID int64, file io.Reader) ([]PlayerResource, error) {
	var player_resources [10]PlayerResource
	maxTicks := 150000

	for i := 0; i < 10; i++ {
		player_resources[i].Slot = i
	}

	p, err := manta.NewStreamParser(file)
	if err != nil {
		return nil, fmt.Errorf("unable to create parser: %s", err)
	}

	var current_tick int = 0
	playerDataFound := false
	heroMapByEntIndex := make(map[uint32]string)
	heroMapByHandle := make(map[uint32]string)
	entIndexToSlot := make(map[uint32]int)
	heroHandles := make(map[int]uint64)
	allHeroEntities := make(map[uint32]string)
	playerSteamIDs := make(map[int]uint64)
	slotsWithPlayers := 0
	slotsWithHeroes := 0

	checkComplete := func() bool {
		if !playerDataFound || slotsWithPlayers == 0 {
			return false
		}
		slotsWithHeroes = 0
		for i := 0; i < 10; i++ {
			if player_resources[i].Name != "" {
				if player_resources[i].Hero != "" {
					slotsWithHeroes++
				}
			}
		}
		return slotsWithHeroes >= slotsWithPlayers
	}

	p.Callbacks.OnCNETMsg_Tick(func(m *dota.CNETMsg_Tick) error {
		current_tick = int(m.GetTick())
		if current_tick > maxTicks && !playerDataFound {
			return fmt.Errorf("timeout: player data not found within %d ticks", maxTicks)
		}
		if checkComplete() {
			return fmt.Errorf("early_stop_complete")
		}
		return nil
	})

	p.OnEntity(func(e *manta.Entity, op manta.EntityOp) error {
		className := e.GetClassName()
		entIndex := uint32(e.GetIndex())

		if className == "CDOTA_PlayerResource" {
			playerDataFound = true
			for i := 0; i < 10; i++ {
				player_resources[i].Slot = i

				if steamid, steamidok := e.GetUint64(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerSteamID", i)); steamidok {
					player_resources[i].SteamID = steamid
					if steamid > 0 {
						playerSteamIDs[i] = steamid
					}
				}

				if entindex, entindexok := e.GetUint32(fmt.Sprintf("m_vecPlayerData.000%d.m_nPlayerSlot", i)); entindexok {
					player_resources[i].EntIndex = entindex
					entIndexToSlot[entindex] = i
				}

				if team, teamok := e.GetInt32(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerTeam", i)); teamok {
					player_resources[i].Team = team
				}

				if name, nameok := e.GetString(fmt.Sprintf("m_vecPlayerData.000%d.m_iszPlayerName", i)); nameok {
					if player_resources[i].Name == "" && name != "" {
						slotsWithPlayers++
					}
					player_resources[i].Name = name
				}

				if heroHandle64, heroHandleOk := e.GetUint64(fmt.Sprintf("m_vecPlayerTeamData.000%d.m_hSelectedHero", i)); heroHandleOk {
					if heroHandle64 != 0 && heroHandle64 != 16777215 {
						heroHandles[i] = heroHandle64
					}
				}
			}
		}

		if className == "CDOTAPlayerController" {
			if playerID, ok := e.GetInt32("m_nPlayerID"); ok && playerID >= 0 && playerID < 10 {
				if entindex, ok := e.GetUint32("m_nPlayerSlot"); ok {
					entIndexToSlot[entindex] = int(playerID)
				}
			}
		}

		if strings.HasPrefix(className, "CDOTA_Unit_Hero_") {
			heroName := strings.TrimPrefix(className, "CDOTA_Unit_Hero_")
			heroMapByHandle[entIndex] = heroName
			allHeroEntities[entIndex] = heroName

			if playerID, ok := e.GetInt32("m_iPlayerID"); ok && playerID >= 0 && playerID < 10 {
				if player_resources[playerID].Hero == "" {
					player_resources[playerID].Hero = heroName
				}
			}

			if owner, ok := e.GetUint32("m_hOwnerEntity"); ok && owner != 0 {
				heroMapByEntIndex[owner] = heroName
				if slot, ok := entIndexToSlot[owner]; ok {
					if player_resources[slot].Hero == "" {
						player_resources[slot].Hero = heroName
					}
				}
			}

			if owner, ok := e.GetUint32("m_hOwner"); ok && owner != 0 {
				heroMapByEntIndex[owner] = heroName
				if slot, ok := entIndexToSlot[owner]; ok {
					if player_resources[slot].Hero == "" {
						player_resources[slot].Hero = heroName
					}
				}
			}

			if heroSteamID, ok := e.GetUint64("m_steamID"); ok && heroSteamID > 0 {
				for slot, steamID := range playerSteamIDs {
					if steamID == heroSteamID {
						if player_resources[slot].Hero == "" {
							player_resources[slot].Hero = heroName
						}
					}
				}
			}
		}

		return nil
	})

	parseError := p.Start()
	if parseError != nil {
		if parseError.Error() == "early_stop_complete" {
		} else {
			return nil, fmt.Errorf("parser error: %v", parseError)
		}
	}

	for i := 0; i < 10; i++ {
		if player_resources[i].Hero == "" {
			if heroHandle64, ok := heroHandles[i]; ok {
				if heroEntity := p.FindEntityByHandle(heroHandle64); heroEntity != nil {
					heroClassName := heroEntity.GetClassName()
					if strings.HasPrefix(heroClassName, "CDOTA_Unit_Hero_") {
						heroName := strings.TrimPrefix(heroClassName, "CDOTA_Unit_Hero_")
						player_resources[i].Hero = heroName
					}
				}
				if player_resources[i].Hero == "" {
					entityIndexFromHandle := uint32(heroHandle64) & 0x7FFF
					if heroName, ok := heroMapByHandle[entityIndexFromHandle]; ok {
						player_resources[i].Hero = heroName
					}
				}
				if player_resources[i].Hero == "" {
					if heroName, ok := heroMapByHandle[uint32(heroHandle64)]; ok {
						player_resources[i].Hero = heroName
					}
				}
				if player_resources[i].Hero == "" {
					entityIndexFromHandle := uint32(heroHandle64) & 0x7FFF
					for entIdx, heroName := range allHeroEntities {
						decodedIdx := uint32(entIdx) & 0x7FFF
						if decodedIdx == entityIndexFromHandle || uint32(entIdx) == entityIndexFromHandle {
							player_resources[i].Hero = heroName
							break
						}
					}
				}
			}
			if player_resources[i].Hero == "" && player_resources[i].EntIndex > 0 {
				if hero, ok := heroMapByEntIndex[player_resources[i].EntIndex]; ok {
					player_resources[i].Hero = hero
				}
			}
		}
	}

	result := make([]PlayerResource, 10)
	copy(result[:], player_resources[:])

	return result, nil
}

func ParseReplay(matchID int64, file io.Reader, reportedSlot int, reportedSteamID uint64) (ParseResult, error) {
	fmt.Printf("[PARSER] Starting ParseReplay - matchID: %d, reportedSlot: %d, reportedSteamID: %d\n", matchID, reportedSlot, reportedSteamID)

	var player_resources [10]PlayerResource

	var teamReports int = 0
	var enemyReports int = 0

	var current_tick int = 0
	var begin_tick int = 0
	var scoreboardOpen map[uint64]bool //steamid as index

	var reportedTeam int = 2
	var pausedTicks int = 0

	var reports []*Report

	entityCounts := make(map[string]int)

	lastHoverTick := make(map[int]int)    // reporter slot -> tick when hovered
	lastHoverTarget := make(map[int]int)  // reporter slot -> target slot being hovered
	firstHoverTarget := make(map[int]int) // reporter slot -> first target slot hovered (before moving cursor)

	heroMapByEntIndex := make(map[uint32]string)
	heroMapByHandle := make(map[uint32]string)
	entIndexToSlot := make(map[uint32]int)
	heroHandles := make(map[int]uint64)
	allHeroEntities := make(map[uint32]string)
	playerSteamIDs := make(map[int]uint64)

	// Use local variables that can be modified (shadow the parameters)
	// These need to be modifiable within closures
	actualReportedSlot := reportedSlot
	actualReportedSteamID := reportedSteamID
	reportedPlayerFound := false

	// If SteamID is provided, reset slot to -1 to force lookup by SteamID
	// This prevents accidental default to slot 0 if reportedSlot was passed as 0 (default int)
	if actualReportedSteamID > 0 {
		actualReportedSlot = -1
		fmt.Printf("[PARSER] SteamID provided, resetting slot to -1 for lookup\n")
	}

	// If only SteamID is provided, we'll find the slot later
	// If only slot is provided, we'll find the SteamID later

	// Initialize the map
	scoreboardOpen = make(map[uint64]bool)

	fmt.Printf("[PARSER] Creating stream parser...\n")
	p, err := manta.NewStreamParser(file)
	if err != nil {
		fmt.Printf("[PARSER] ERROR: Failed to create parser: %s\n", err)
		return ParseResult{}, fmt.Errorf("unable to create parser: %s", err)
	}
	fmt.Printf("[PARSER] Parser created successfully\n")

	p.Callbacks.OnCNETMsg_Tick(func(m *dota.CNETMsg_Tick) error {
		current_tick = int(m.GetTick())
		return nil
	})

	p.Callbacks.OnCDOTAUserMsg_GamerulesStateChanged(func(m *dota.CDOTAUserMsg_GamerulesStateChanged) error {
		if m.GetState() == 5 {
			begin_tick = current_tick
			fmt.Printf("[PARSER] Game started! begin_tick set to: %d\n", begin_tick)
		}
		return nil
	})

	p.OnEntity(func(e *manta.Entity, op manta.EntityOp) error {
		className := e.GetClassName()
		entityCounts[className]++

		if className == "CDOTA_PlayerResource" {
			for i := 0; i < 10; i++ {
				isVictim := false
				if actualReportedSlot != -1 && i == actualReportedSlot {
					isVictim = true
				}

				if steamid, steamidok := e.GetUint64(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerSteamID", i)); steamidok {
					player_resources[i].SteamID = steamid
					if steamid > 0 {
						playerSteamIDs[i] = steamid
					}

					if actualReportedSteamID > 0 {
						if steamid == actualReportedSteamID {
							isVictim = true
							if actualReportedSlot != i {
								actualReportedSlot = i
								if !reportedPlayerFound {
									fmt.Printf("[PARSER] Found reported player by SteamID! Slot: %d, SteamID: %d\n", i, steamid)
									reportedPlayerFound = true
								}
							}
						}
					} else if isVictim {
						if actualReportedSteamID != steamid {
							actualReportedSteamID = steamid
							if !reportedPlayerFound {
								fmt.Printf("[PARSER] Found reported player by slot! Slot: %d, SteamID: %d\n", i, steamid)
								reportedPlayerFound = true
							}
						}
					}
				}

				if entindex, entindexok := e.GetUint32(fmt.Sprintf("m_vecPlayerData.000%d.m_nPlayerSlot", i)); entindexok {
					player_resources[i].EntIndex = entindex
					entIndexToSlot[entindex] = i
				}

				if team, teamok := e.GetInt32(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerTeam", i)); teamok {
					player_resources[i].Team = team
					if isVictim {
						reportedTeam = int(team)
					}
				}

				if name, nameok := e.GetString(fmt.Sprintf("m_vecPlayerData.000%d.m_iszPlayerName", i)); nameok {
					player_resources[i].Name = name
				}

				if heroHandle64, heroHandleOk := e.GetUint64(fmt.Sprintf("m_vecPlayerTeamData.000%d.m_hSelectedHero", i)); heroHandleOk {
					if heroHandle64 != 0 && heroHandle64 != 16777215 {
						heroHandles[i] = heroHandle64
					}
				}
			}

		} else if className == "CDOTAGamerulesProxy" {
			if v, ok := e.GetInt32("m_pGameRules.m_nTotalPausedTicks"); ok {
				if v > 0 {
					pausedTicks = int(v)
				}
			}
		}

		if className == "CDOTAPlayerController" {
			if playerID, ok := e.GetInt32("m_nPlayerID"); ok && playerID >= 0 && playerID < 10 {
				if entindex, ok := e.GetUint32("m_nEntityIndex"); ok {
					entIndexToSlot[entindex] = int(playerID)
				}
			}
		}

		if strings.HasPrefix(className, "CDOTA_Unit_Hero_") {
			heroName := strings.TrimPrefix(className, "CDOTA_Unit_Hero_")
			entIndex := uint32(e.GetIndex())
			heroMapByHandle[entIndex] = heroName
			allHeroEntities[entIndex] = heroName

			if playerID, ok := e.GetInt32("m_iPlayerID"); ok && playerID >= 0 && playerID < 10 {
				if player_resources[playerID].Hero == "" {
					player_resources[playerID].Hero = heroName
				}
			}

			if owner, ok := e.GetUint32("m_hOwnerEntity"); ok && owner != 0 {
				heroMapByEntIndex[owner] = heroName
				if slot, ok := entIndexToSlot[owner]; ok {
					if player_resources[slot].Hero == "" {
						player_resources[slot].Hero = heroName
					}
				}
			}

			if owner, ok := e.GetUint32("m_hOwner"); ok && owner != 0 {
				heroMapByEntIndex[owner] = heroName
				if slot, ok := entIndexToSlot[owner]; ok {
					if player_resources[slot].Hero == "" {
						player_resources[slot].Hero = heroName
					}
				}
			}

			if heroSteamID, ok := e.GetUint64("m_steamID"); ok && heroSteamID > 0 {
				for slot, steamID := range playerSteamIDs {
					if steamID == heroSteamID {
						if player_resources[slot].Hero == "" {
							player_resources[slot].Hero = heroName
						}
					}
				}
			}
		}

		if begin_tick == 0 {
			if className != "CDOTA_PlayerResource" && className != "CDOTAGamerulesProxy" {
				return nil
			}
		}
		if className == "CDOTAPlayerController" {
			if steamid, ok2 := e.GetUint64("m_steamID"); ok2 {
				if name, ok3 := e.GetString("m_iszPlayerName"); ok3 {
					parseAllReports := (actualReportedSlot == -1 && actualReportedSteamID == 0)
					if parseAllReports || steamid != actualReportedSteamID {
						if statsPanel, ok := e.GetInt32("m_iStatsPanel"); ok {
							if statsPanel == 1 {
								if !scoreboardOpen[steamid] {
								}
								if xpos, xposok := e.GetInt32("m_iCursor.0000"); xposok {
									if ypos, yposok := e.GetInt32("m_iCursor.0001"); yposok {
										xpos = int32(math.Round(float64(xpos) / 510 * 1920))
										ypos = int32(math.Round(float64(ypos) / 383 * 1080))

										if aspect, aspectok := e.GetFloat32("m_flAspectRatio"); aspectok {
											targetSlot := isReportButton(int(xpos), int(ypos), aspect)

											for i := 0; i < 10; i++ {
												if player_resources[i].SteamID == steamid {
													inConfirmBox := xpos >= 956 && xpos <= 1170 && ypos >= 847 && ypos <= 888

													if inConfirmBox {
														if hoverTick, exists := lastHoverTick[i]; exists {
															tickDiff := current_tick - hoverTick
															if tickDiff >= 0 && tickDiff <= 120 {
																firstTargetSlot, hasFirstTarget := firstHoverTarget[i]
																lastTargetSlot, hasLastTarget := lastHoverTarget[i]

																finalTargetSlot := firstTargetSlot
																if !hasFirstTarget || firstTargetSlot < 0 || firstTargetSlot >= 10 || firstTargetSlot == i {
																	if hasLastTarget && lastTargetSlot >= 0 && lastTargetSlot < 10 && lastTargetSlot != i {
																		fmt.Printf("[PARSER] WARNING: First targetSlot %d is invalid/missing/self, using last targetSlot %d\n", firstTargetSlot, lastTargetSlot)
																		finalTargetSlot = lastTargetSlot
																	} else if targetSlot >= 0 && targetSlot < 10 && targetSlot != i {
																		fmt.Printf("[PARSER] WARNING: First and last targetSlots invalid/self, using current targetSlot %d\n", targetSlot)
																		finalTargetSlot = targetSlot
																	} else {
																		fmt.Printf("[PARSER] ERROR: All targetSlots are invalid/self (first: %d, last: %d, current: %d, reporter: %d), skipping\n", firstTargetSlot, lastTargetSlot, targetSlot, i)
																		continue
																	}
																}

																if finalTargetSlot == i {
																	fmt.Printf("[PARSER] ERROR: Final targetSlot %d is the same as reporter slot %d (self-report), skipping\n", finalTargetSlot, i)
																	continue
																}

																if finalTargetSlot != firstTargetSlot && hasFirstTarget {
																	fmt.Printf("[PARSER] WARNING: Using first targetSlot %d but it differs from stored first %d. This might indicate a bug.\n", finalTargetSlot, firstTargetSlot)
																}

																if finalTargetSlot >= 0 && finalTargetSlot < 10 {
																	targetSteamID := player_resources[finalTargetSlot].SteamID
																	targetName := player_resources[finalTargetSlot].Name
																	targetHero := player_resources[finalTargetSlot].Hero

																	if targetSteamID == 0 {
																		fmt.Printf("[PARSER] WARNING: Target slot %d has SteamID 0, trying to find SteamID from playerSteamIDs map\n", finalTargetSlot)
																		if foundSteamID, exists := playerSteamIDs[finalTargetSlot]; exists && foundSteamID > 0 {
																			targetSteamID = foundSteamID
																			fmt.Printf("[PARSER] Found SteamID %d for slot %d from playerSteamIDs map\n", targetSteamID, finalTargetSlot)
																		} else {
																			fmt.Printf("[PARSER] ERROR: Could not find SteamID for target slot %d, skipping report\n", finalTargetSlot)
																			continue
																		}
																	}

																	minutes, secs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, hoverTick)

																	var reportTeam string
																	if team, okteam := e.GetUint64("m_iTeamNum"); okteam {
																		reporterTeam := int(team)
																		targetTeam := player_resources[finalTargetSlot].Team
																		if targetTeam == int32(reporterTeam) {
																			reportTeam = "FRIENDLY"
																			teamReports += 1
																		} else {
																			reportTeam = "ENEMY"
																			enemyReports += 1
																		}
																	}

																	newReport := &Report{
																		Time:          fmt.Sprintf("%02d:%02d", minutes, secs),
																		SteamID:       steamid,
																		Slot:          i,
																		Name:          name,
																		Team:          reportTeam,
																		Hero:          player_resources[i].Hero,
																		TargetSlot:    finalTargetSlot,
																		TargetSteamID: targetSteamID,
																		TargetName:    targetName,
																		TargetHero:    targetHero,
																	}

																	reports = append(reports, newReport)
																	delete(lastHoverTick, i)
																	delete(lastHoverTarget, i)
																	delete(firstHoverTarget, i)
																}
															}
														}
													}

													if targetSlot != -1 && targetSlot != i {
														firstTarget, hadFirstTarget := firstHoverTarget[i]
														if !hadFirstTarget || firstTarget < 0 || firstTarget >= 10 || firstTarget == i {
															firstHoverTarget[i] = targetSlot
														}
														lastHoverTick[i] = current_tick
														lastHoverTarget[i] = targetSlot
													}
													break
												}
											}
										}
									}
								}
								scoreboardOpen[steamid] = true
							} else if statsPanel == 0 {
								scoreboardOpen[steamid] = false
							}
						}
					}
				}
			}
		}

		return nil
	})

	fmt.Printf("[PARSER] Starting parser execution...\n")
	fmt.Printf("[PARSER] Current state before Start - begin_tick: %d, current_tick: %d\n", begin_tick, current_tick)

	packetEntityCount := 0
	packetCount := 0
	lastPacketTick := uint32(0)
	demoMessageCount := 0

	p.Callbacks.OnCDemoFileHeader(func(m *dota.CDemoFileHeader) error {
		demoMessageCount++
		return nil
	})

	p.Callbacks.OnCDemoFileInfo(func(m *dota.CDemoFileInfo) error {
		demoMessageCount++
		return nil
	})

	p.Callbacks.OnCDemoSendTables(func(m *dota.CDemoSendTables) error {
		demoMessageCount++
		return nil
	})

	p.Callbacks.OnCDemoClassInfo(func(m *dota.CDemoClassInfo) error {
		demoMessageCount++
		return nil
	})

	p.Callbacks.OnCDemoStringTables(func(m *dota.CDemoStringTables) error {
		demoMessageCount++
		return nil
	})

	var lastCDemoPacketDataSize int = 0
	var lastCDemoPacketTick uint32 = 0

	p.Callbacks.OnCDemoPacket(func(m *dota.CDemoPacket) error {
		packetCount++
		dataSize := len(m.GetData())
		lastCDemoPacketDataSize = dataSize
		lastCDemoPacketTick = p.Tick

		if dataSize == 0 {
			fmt.Printf("[PARSER] WARNING: CDemoPacket has empty data buffer at tick %d\n", p.Tick)
		}

		lastPacketTick = p.Tick
		return nil
	})

	p.Callbacks.OnCDemoFullPacket(func(m *dota.CDemoFullPacket) error {
		return nil
	})

	var lastPacketEntityTick uint32 = 0
	var lastPacketEntityBufferSize int = 0
	var lastPacketEntityUpdatedEntries int32 = 0
	var lastPacketEntityMaxEntries int32 = 0

	p.Callbacks.OnCSVCMsg_PacketEntities(func(m *dota.CSVCMsg_PacketEntities) error {
		packetEntityCount++
		entityData := m.GetEntityData()
		bufferSize := len(entityData)
		updatedEntries := m.GetUpdatedEntries()
		serverTick := m.GetServerTick()

		maxEntries := m.GetMaxEntries()
		lastPacketEntityTick = serverTick
		lastPacketEntityBufferSize = bufferSize
		lastPacketEntityUpdatedEntries = updatedEntries
		lastPacketEntityMaxEntries = maxEntries

		if bufferSize == 0 && updatedEntries > 0 {
			fmt.Printf("[PARSER] WARNING: PacketEntities has updatedEntries=%d but empty buffer at tick %d\n", updatedEntries, serverTick)
		}

		if updatedEntries > 0 {
			minExpectedSize := int(updatedEntries) * 2
			if bufferSize < minExpectedSize {
				fmt.Printf("[PARSER] WARNING: PacketEntities buffer may be too small - updatedEntries=%d, bufferSize=%d, minExpectedSize=%d at tick %d\n",
					updatedEntries, bufferSize, minExpectedSize, serverTick)
			}

			if bufferSize < int(updatedEntries) {
				fmt.Printf("[PARSER] CRITICAL: PacketEntities buffer is smaller than entry count - updatedEntries=%d, bufferSize=%d at tick %d. Buffer underflow likely!\n",
					updatedEntries, bufferSize, serverTick)
			}
		}

		return nil
	})

	var parseError error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					parseError = e
				} else {
					parseError = fmt.Errorf("panic: %v", r)
				}
			}
		}()
		parseError = p.Start()
	}()

	if parseError != nil {
		errMsg := parseError.Error()
		fmt.Printf("[PARSER] ERROR: Parser.Start() returned error: %v\n", parseError)
		fmt.Printf("[PARSER] ERROR context - demoMessageCount: %d, lastPacketTick: %d, packetCount: %d, packetEntityCount: %d, current_tick: %d, begin_tick: %d\n",
			demoMessageCount, lastPacketTick, packetCount, packetEntityCount, current_tick, begin_tick)

		if lastCDemoPacketTick > 0 {
			fmt.Printf("[PARSER] ERROR - Last CDemoPacket context - tick: %d, dataSize: %d\n",
				lastCDemoPacketTick, lastCDemoPacketDataSize)
		}

		if lastPacketEntityTick > 0 {
			fmt.Printf("[PARSER] ERROR - Last PacketEntity context - tick: %d, updatedEntries: %d, maxEntries: %d, bufferSize: %d\n",
				lastPacketEntityTick, lastPacketEntityUpdatedEntries, lastPacketEntityMaxEntries, lastPacketEntityBufferSize)
		}

		if strings.Contains(errMsg, "insufficient buffer") {
			contextParts := []string{}
			if lastCDemoPacketTick > 0 {
				contextParts = append(contextParts, fmt.Sprintf("last CDemoPacket: tick=%d, dataSize=%d", lastCDemoPacketTick, lastCDemoPacketDataSize))
			}
			if lastPacketEntityTick > 0 {
				contextParts = append(contextParts, fmt.Sprintf("last PacketEntity: tick=%d, updatedEntries=%d, bufferSize=%d", lastPacketEntityTick, lastPacketEntityUpdatedEntries, lastPacketEntityBufferSize))
			}
			contextStr := strings.Join(contextParts, "; ")
			if contextStr == "" {
				contextStr = "no packet context available"
			}
			enhancedErr := fmt.Errorf("buffer underflow error at tick %d: %v (%s, packetCount=%d, packetEntityCount=%d). This may indicate a corrupted or truncated replay file.",
				current_tick, parseError, contextStr, packetCount, packetEntityCount)
			return ParseResult{}, enhancedErr
		}

		return ParseResult{}, fmt.Errorf("parser error at tick %d: %v", current_tick, parseError)
	}

	for i := 0; i < 10; i++ {
		if player_resources[i].Hero == "" {
			if heroHandle64, ok := heroHandles[i]; ok {
				if heroEntity := p.FindEntityByHandle(heroHandle64); heroEntity != nil {
					heroClassName := heroEntity.GetClassName()
					if strings.HasPrefix(heroClassName, "CDOTA_Unit_Hero_") {
						heroName := strings.TrimPrefix(heroClassName, "CDOTA_Unit_Hero_")
						player_resources[i].Hero = heroName
					}
				}
				if player_resources[i].Hero == "" {
					entityIndexFromHandle := uint32(heroHandle64) & 0x7FFF
					if heroName, ok := heroMapByHandle[entityIndexFromHandle]; ok {
						player_resources[i].Hero = heroName
					}
				}
				if player_resources[i].Hero == "" {
					if heroName, ok := heroMapByHandle[uint32(heroHandle64)]; ok {
						player_resources[i].Hero = heroName
					}
				}
				if player_resources[i].Hero == "" {
					entityIndexFromHandle := uint32(heroHandle64) & 0x7FFF
					for entIdx, heroName := range allHeroEntities {
						decodedIdx := uint32(entIdx) & 0x7FFF
						if decodedIdx == entityIndexFromHandle || uint32(entIdx) == entityIndexFromHandle {
							player_resources[i].Hero = heroName
							break
						}
					}
				}
			}
			if player_resources[i].Hero == "" && player_resources[i].EntIndex > 0 {
				if hero, ok := heroMapByEntIndex[player_resources[i].EntIndex]; ok {
					player_resources[i].Hero = hero
				}
			}
		}
	}

	fmt.Printf("[PARSER] Parser execution completed - begin_tick: %d, current_tick: %d, packetEntityCount: %d\n", begin_tick, current_tick, packetEntityCount)
	fmt.Printf("[PARSER] Final results - TeamReports: %d, EnemyReports: %d, TotalReports: %d\n",
		teamReports, enemyReports, len(reports))
	fmt.Printf("[PARSER] Reported player - Slot: %d, SteamID: %d, Team: %d\n", actualReportedSlot, actualReportedSteamID, reportedTeam)
	fmt.Printf("[PARSER] Game state - begin_tick: %d, pausedTicks: %d, final_tick: %d\n", begin_tick, pausedTicks, current_tick)

	for _, report := range reports {
		if report.Hero == "" && report.Slot >= 0 && report.Slot < 10 {
			report.Hero = player_resources[report.Slot].Hero
		}
	}

	return ParseResult{
		MatchID:      matchID,
		TeamReports:  teamReports,
		EnemyReports: enemyReports,
		Reports:      reports,
	}, nil
}

// GetReplayDate extracts the match date from the replay file header/summary.
// This is extremely fast as it jumps to the footer directly.
func GetReplayDate(file io.Reader) (time.Time, error) {
	// We need a ReadSeeker to jump to the footer.
	rs, ok := file.(io.ReadSeeker)
	if !ok {
		return time.Time{}, fmt.Errorf("file must be an io.ReadSeeker to parse header")
	}

	// Read header (16 bytes)
	header := make([]byte, 16)
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return time.Time{}, fmt.Errorf("failed to seek start: %v", err)
	}
	if _, err := io.ReadFull(rs, header); err != nil {
		return time.Time{}, fmt.Errorf("failed to read header: %v", err)
	}

	// Check magic
	if string(header[0:8]) != "PBDEMS2\000" {
		// Try fallback magic or just error
		// Some might be different? But standard is PBDEMS2\0
	}

	offset1 := binary.LittleEndian.Uint32(header[8:12])

	// Check file size to ensure offset is valid
	endPos, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to seek end: %v", err)
	}

	if int64(offset1) >= endPos {
		return time.Time{}, fmt.Errorf("invalid offset in header")
	}

	// Seek to offset1
	if _, err := rs.Seek(int64(offset1), io.SeekStart); err != nil {
		return time.Time{}, fmt.Errorf("failed to seek to offset1: %v", err)
	}

	// Read Cmd (varint)
	br := &byteReader{r: rs}
	cmd, err := binary.ReadUvarint(br)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read cmd: %v", err)
	}

	isCompressed := (cmd & 0x40) != 0
	// cmd = cmd &^ 0x40 // Base cmd, e.g. DEM_FileInfo = 2

	// Read Tick (varint)
	_, err = binary.ReadUvarint(br)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read tick: %v", err)
	}

	// Read Size (varint)
	size, err := binary.ReadUvarint(br)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read size: %v", err)
	}

	data := make([]byte, size)
	if _, err := io.ReadFull(rs, data); err != nil {
		return time.Time{}, fmt.Errorf("failed to read data: %v", err)
	}

	if isCompressed {
		decoded, err := snappy.Decode(nil, data)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to decode snappy: %v", err)
		}
		data = decoded
	}

	// Unmarshal CDemoFileInfo
	info := &dota.CDemoFileInfo{}
	if err := proto.Unmarshal(data, info); err != nil {
		return time.Time{}, fmt.Errorf("failed to unmarshal CDemoFileInfo: %v", err)
	}

	if info.GameInfo != nil && info.GameInfo.Dota != nil {
		endTime := info.GameInfo.Dota.GetEndTime()
		if endTime > 0 {
			return time.Unix(int64(endTime), 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("end_time not found in GameInfo")
}

type byteReader struct {
	r io.Reader
}

func (b *byteReader) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	_, err := b.r.Read(buf)
	return buf[0], err
}
