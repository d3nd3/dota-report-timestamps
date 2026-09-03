package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
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

type Candidate struct {
	Slot    int     `json:"Slot"`
	SteamID uint64  `json:"SteamID"`
	Name    string  `json:"Name"`
	Hero    string  `json:"Hero"`
	Score   float64 `json:"Score"`
}

type Report struct {
	Time                   string      `json:"Time"`
	Team                   string      `json:"Team"` // "FRIENDLY" or "ENEMY"
	SteamID                uint64      `json:"SteamID"`
	Slot                   int         `json:"Slot"`
	Name                   string      `json:"Name"`
	Hero                   string      `json:"Hero"`
	TargetSlot             int         `json:"TargetSlot"`                       // The slot of the player who was reported
	TargetSteamID          uint64      `json:"TargetSteamID"`                    // The SteamID of the player who was reported
	TargetName             string      `json:"TargetName"`                       // The name of the player who was reported
	TargetHero             string      `json:"TargetHero"`                       // The hero of the player who was reported
	Candidates             []Candidate `json:"Candidates,omitempty"`             // Top candidates with scores
	SelectedCandidateIndex int         `json:"SelectedCandidateIndex,omitempty"` // Index of selected candidate
}

type TimeRange struct {
	Start string `json:"Start"` // Format: "MM:SS"
	End   string `json:"End"`   // Format: "MM:SS"
}

type ScoreboardActivity struct {
	SteamID    uint64      `json:"SteamID"`
	Slot       int         `json:"Slot"`
	Name       string      `json:"Name"`
	Hero       string      `json:"Hero"`
	Team       int32       `json:"Team"` // 2 = radiant, 3 = dire
	TimeRanges []TimeRange `json:"TimeRanges"`
}

type ParseResult struct {
	MatchID              int64                 `json:"MatchID"`
	TeamReports          int                   `json:"TeamReports"`
	EnemyReports         int                   `json:"EnemyReports"`
	Reports              []*Report             `json:"Reports"`
	ScoreboardActivities []*ScoreboardActivity `json:"scoreboardActivities"`
	Players              []PlayerResource      `json:"Players"`
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
    // Width of their scoreboard in 1920 pixels, based on their custom resolution, aspectRatio.
    score_width := int(math.Round(float64((1.77777777778 * (920.0 / 1920)) / float64(aspect) * 1920)))
    score_width_no_tips := int(math.Round(float64((1.77777777778 * (820.0 / 1920)) / float64(aspect) * 1920)))

	// 6 good mb
    padding := 6
    paddingFloat := float64(padding)

    lower_x_bound := 866.0 - paddingFloat
    upper_x_bound := 893.0 + paddingFloat

    lower_x_bound_r := lower_x_bound / 920
    upper_x_bound_r := upper_x_bound / 920

    lower_x_bound_no_tips_r := (lower_x_bound - 100) / 820
    upper_x_bound_no_tips_r := (upper_x_bound - 100) / 820

    if (x >= int(math.Floor(float64(lower_x_bound_no_tips_r*float64(score_width_no_tips)))) && x <= int(math.Ceil(float64(upper_x_bound_no_tips_r*float64(score_width_no_tips))))) || (x >= int(math.Floor(float64(lower_x_bound_r*float64(score_width)))) && x <= int(math.Ceil(float64(upper_x_bound_r*float64(score_width))))) {
        if y >= 106-padding && y <= 135+padding {
            return 0
        } else if y >= 176-padding && y <= 205+padding {
            return 1
        } else if y >= 246-padding && y <= 275+padding {
            return 2
        } else if y >= 316-padding && y <= 345+padding {
            return 3
        } else if y >= 386-padding && y <= 415+padding {
            return 4
        } else if y >= 488-padding && y <= 517+padding {
            return 5
        } else if y >= 558-padding && y <= 587+padding {
            return 6
        } else if y >= 628-padding && y <= 657+padding {
            return 7
        } else if y >= 698-padding && y <= 727+padding {
            return 8
        } else if y >= 768-padding && y <= 797+padding {
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
	for i := 0; i < 10; i++ {
		player_resources[i].Slot = i
	}

	var teamReports int = 0
	var enemyReports int = 0

	var current_tick int = 0
	var begin_tick int = 0
	const minHoverDuration = 1                           // Minimum hover duration in ticks for a session to be considered valid
	const validReportWindowSeconds = 4                   // Maximum time window in seconds between reportCircle and confirmButton
	const durationWeight = 0.7                           // Weight for duration in report attribution scoring
	const recencyWeight = 0.3                            // Weight for recency in report attribution scoring
	const ticksPerSecond = 30                            // Game tick rate (ticks per second)
	const scoreboardGracePeriodTicks = 11               // 300ms grace period after scoreboard closes (300ms * 30 ticks/sec = 9 ticks)
	scoreboardOpenStartTick := make(map[uint64]int)      // steamid -> tick when scoreboard opened
	scoreboardTimeRanges := make(map[uint64][]TimeRange) // steamid -> list of time ranges
	lastScoreboardOpenTick := make(map[uint64]int)       // steamid -> tick when scoreboard was last open

	var reportedTeam int = 2
	var pausedTicks int = 0

	var reports []*Report

	entityCounts := make(map[string]int)

	hoverDurations := make(map[int]map[int]int)            // reporter slot -> target slot -> max session duration in ticks
	lastHoverTime := make(map[int]int)                     // reporter slot -> last tick any report button was hovered
	targetLastHoverTime := make(map[int]map[int]int)       // reporter slot -> target slot -> last tick this target was hovered
	targetLastValidHoverTime := make(map[int]map[int]int)  // reporter slot -> target slot -> last tick a valid (>=min) session ended
	targetFirstValidHoverTime := make(map[int]map[int]int) // reporter slot -> target slot -> first tick a valid (>=min) session ended
	hoverSessionStart := make(map[int]map[int]int)         // reporter slot -> target slot -> tick when current hover session started
	hoverSessionDurations := make(map[int]map[int][]int)   // reporter slot -> target slot -> list of completed session durations
	confirmBoxHoverStart := make(map[int]int)              // reporter slot -> tick when confirm box hover started

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
					if statsPanel, ok := e.GetInt32("m_iStatsPanel"); ok {
						if statsPanel == 1 {
							lastScoreboardOpenTick[steamid] = current_tick
							if _, exists := scoreboardOpenStartTick[steamid]; !exists {
								scoreboardOpenStartTick[steamid] = current_tick
							}
						} else if statsPanel == 0 {
							if startTick, exists := scoreboardOpenStartTick[steamid]; exists {
								if begin_tick > 0 {
									startMinutes, startSecs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, startTick)
									endMinutes, endSecs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, current_tick)
									timeRange := TimeRange{
										Start: fmt.Sprintf("%02d:%02d", startMinutes, startSecs),
										End:   fmt.Sprintf("%02d:%02d", endMinutes, endSecs),
									}
									scoreboardTimeRanges[steamid] = append(scoreboardTimeRanges[steamid], timeRange)
								}
								delete(scoreboardOpenStartTick, steamid)
							}
						}
					}

					parseAllReports := (actualReportedSlot == -1 && actualReportedSteamID == 0)
					if parseAllReports || steamid != actualReportedSteamID {
						statsPanel, statsPanelOk := e.GetInt32("m_iStatsPanel")
						if xpos, xposok := e.GetInt32("m_iCursor.0000"); xposok {
							if ypos, yposok := e.GetInt32("m_iCursor.0001"); yposok {
								xpos = int32(math.Round(float64(xpos) / 510 * 1920))
								ypos = int32(math.Round(float64(ypos) / 383 * 1080))

								if aspect, aspectok := e.GetFloat32("m_flAspectRatio"); aspectok {
									targetSlot := isReportButton(int(xpos), int(ypos), aspect)

									for i := 0; i < 10; i++ {
										if player_resources[i].SteamID == steamid {
											// Initialize hover maps for this reporter if needed
											if _, ok := hoverDurations[i]; !ok {
												hoverDurations[i] = make(map[int]int)
											}
											if _, ok := targetLastHoverTime[i]; !ok {
												targetLastHoverTime[i] = make(map[int]int)
											}
											if _, ok := targetLastValidHoverTime[i]; !ok {
												targetLastValidHoverTime[i] = make(map[int]int)
											}
											if _, ok := targetFirstValidHoverTime[i]; !ok {
												targetFirstValidHoverTime[i] = make(map[int]int)
											}
											if _, ok := hoverSessionStart[i]; !ok {
												hoverSessionStart[i] = make(map[int]int)
											}
											if _, ok := hoverSessionDurations[i]; !ok {
												hoverSessionDurations[i] = make(map[int][]int)
											}

											// Track hover sessions (when scoreboard is open or within 300ms grace period after closing)
											isScoreboardOpen := statsPanelOk && statsPanel == 1
											if !isScoreboardOpen {
												if lastOpenTick, exists := lastScoreboardOpenTick[steamid]; exists {
													if current_tick-lastOpenTick <= scoreboardGracePeriodTicks {
														isScoreboardOpen = true
													}
												}
											}

											if isScoreboardOpen {
												// scoreboard is open
												fmt.Printf("[PARSER] DEBUG: Slot %d - scoreboard open at tick %d\n", i, current_tick)
												if targetSlot != -1 && targetSlot != i {
													fmt.Printf("[PARSER] DEBUG: Slot %d - hovering over target slot %d at tick %d\n", i, targetSlot, current_tick)
													// End any active sessions for other targets (switching targets)
													for tSlot, sessionStart := range hoverSessionStart[i] {
														if tSlot != targetSlot && sessionStart > 0 {
															sessionDuration := current_tick - sessionStart
															if sessionDuration > 0 {
																hoverSessionDurations[i][tSlot] = append(hoverSessionDurations[i][tSlot], sessionDuration)
																if sessionDuration > hoverDurations[i][tSlot] {
																	hoverDurations[i][tSlot] = sessionDuration
																}
																if sessionDuration >= minHoverDuration {
																	targetLastValidHoverTime[i][tSlot] = current_tick
																	if _, exists := targetFirstValidHoverTime[i][tSlot]; !exists {
																		targetFirstValidHoverTime[i][tSlot] = current_tick
																	}
																}
															}
															delete(hoverSessionStart[i], tSlot)
														}
													}

													// Check if this is a new hover session (not currently hovering this target)
													if _, hasActiveSession := hoverSessionStart[i][targetSlot]; !hasActiveSession {
														// Start new session
														hoverSessionStart[i][targetSlot] = current_tick
														fmt.Printf("[PARSER] DEBUG: Slot %d - new hover session started for target slot %d at tick %d\n", i, targetSlot, current_tick)
													}
													lastHoverTime[i] = current_tick
													targetLastHoverTime[i][targetSlot] = current_tick
												} else {
													// Cursor is not over any report button - end any active sessions
													for tSlot, sessionStart := range hoverSessionStart[i] {
														if sessionStart > 0 {
															sessionDuration := current_tick - sessionStart
															if sessionDuration > 0 {
																hoverSessionDurations[i][tSlot] = append(hoverSessionDurations[i][tSlot], sessionDuration)
																// Update max duration if this session is longer
																if sessionDuration > hoverDurations[i][tSlot] {
																	hoverDurations[i][tSlot] = sessionDuration
																}
																if sessionDuration >= minHoverDuration {
																	targetLastValidHoverTime[i][tSlot] = current_tick
																	if _, exists := targetFirstValidHoverTime[i][tSlot]; !exists {
																		targetFirstValidHoverTime[i][tSlot] = current_tick
																	}
																}
															}
															delete(hoverSessionStart[i], tSlot)
														}
													}
												}
											} else {
												// Scoreboard closed - end any active sessions
												for tSlot, sessionStart := range hoverSessionStart[i] {
													if sessionStart > 0 {
														sessionDuration := current_tick - sessionStart
														if sessionDuration > 0 {
															hoverSessionDurations[i][tSlot] = append(hoverSessionDurations[i][tSlot], sessionDuration)
															if sessionDuration > hoverDurations[i][tSlot] {
																hoverDurations[i][tSlot] = sessionDuration
															}
															if sessionDuration >= minHoverDuration {
																fmt.Printf("[PARSER] DEBUG: Slot %d - valid hover session ended at tick %d (duration: %d ticks)\n", tSlot, current_tick, sessionDuration)
																targetLastValidHoverTime[i][tSlot] = current_tick
																if _, exists := targetFirstValidHoverTime[i][tSlot]; !exists {
																	targetFirstValidHoverTime[i][tSlot] = current_tick
																}
															}
														}
														delete(hoverSessionStart[i], tSlot)
													}
												}
											}

											// Check confirm box (independent of scoreboard state)
											inConfirmBox := xpos >= 968 && xpos <= 1159 && ypos >= 844 && ypos <= 889

											if inConfirmBox {
												// Track confirm box hover start if this is a new hover session
												if _, exists := confirmBoxHoverStart[i]; !exists {
													confirmBoxHoverStart[i] = current_tick
												}
												// Finalize any active hover sessions before calculating scores
												for tSlot, sessionStart := range hoverSessionStart[i] {
													if sessionStart > 0 {
														sessionDuration := current_tick - sessionStart
														if sessionDuration > 0 {
															fmt.Printf("[PARSER] DEBUG: Slot %d - confirm box hover detected, finalizing session (duration: %d ticks)\n", tSlot, sessionDuration)
															hoverSessionDurations[i][tSlot] = append(hoverSessionDurations[i][tSlot], sessionDuration)
															if sessionDuration > hoverDurations[i][tSlot] {
																hoverDurations[i][tSlot] = sessionDuration
															}
															if sessionDuration >= minHoverDuration {
																fmt.Printf("[PARSER] DEBUG: Slot %d - valid hover session ended at tick %d (duration: %d ticks)\n", tSlot, current_tick, sessionDuration)
																targetLastValidHoverTime[i][tSlot] = current_tick
																if _, exists := targetFirstValidHoverTime[i][tSlot]; !exists {
																	targetFirstValidHoverTime[i][tSlot] = current_tick
																}
															}
														}
														delete(hoverSessionStart[i], tSlot)
													}
												}
											} else {
												// Cursor is not in confirmBox
												if startTick, exists := confirmBoxHoverStart[i]; exists {
													// Cursor was in confirmBox but now is not.
													confirmBoxHoverDuration := current_tick - startTick
													minutes, secs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, current_tick)
													timestamp := fmt.Sprintf("%02d:%02d", minutes, secs)

													fmt.Printf("\n[PARSER] =================== Report Detection (Time: %s) ===================\n", timestamp)

													// Check if any overlapping slots (6/8/9) were hovered
													hasOverlappingSlots := false
													for tSlot := range hoverDurations[i] {
														if tSlot == 6 || tSlot == 7 || tSlot == 8 {
															hasOverlappingSlots = true
															fmt.Printf("[PARSER] DEBUG: Overlapping slot detected: %d\n", tSlot)
															break
														}
													}

													var earliestFirstHoverTime int = -1
													var earliestFirstHoverSlot int = -1
													var mostRecentLastHoverTime int = -1
													var mostRecentLastHoverSlot int = -1

													if hasOverlappingSlots {
														// When overlapping slots are present, use first-hover algorithm for ALL slots
														// Find the earliest first hover time across slots that have recent hover sessions AND are within valid window
														fmt.Printf("[PARSER] DEBUG: Checking first-hover algorithm - startTick: %d, validWindow: %d ticks\n", startTick, validReportWindowSeconds*ticksPerSecond)
														for tSlot := range hoverDurations[i] {
															if _, hasRecentSession := hoverDurations[i][tSlot]; hasRecentSession {
																var candidateTime int = -1
																// Try targetFirstValidHoverTime first
																if targetFirstValidHoverTime[i] != nil {
																	if firstTime, exists := targetFirstValidHoverTime[i][tSlot]; exists {
																		tickDiff := startTick - firstTime
																		fmt.Printf("[PARSER] DEBUG: Slot %d - targetFirstValidHoverTime: %d, tickDiff: %d\n", tSlot, firstTime, tickDiff)
																		if tickDiff >= 0 && tickDiff <= validReportWindowSeconds*ticksPerSecond {
																			candidateTime = firstTime
																		}
																	}
																}
																// Fallback to targetLastValidHoverTime if first is missing or outside window
																if candidateTime == -1 && targetLastValidHoverTime[i] != nil {
																	if lastTime, exists := targetLastValidHoverTime[i][tSlot]; exists {
																		tickDiff := startTick - lastTime
																		fmt.Printf("[PARSER] DEBUG: Slot %d - fallback to targetLastValidHoverTime: %d, tickDiff: %d\n", tSlot, lastTime, tickDiff)
																		if tickDiff >= 0 && tickDiff <= validReportWindowSeconds*ticksPerSecond {
																			candidateTime = lastTime
																		}
																	}
																}
																if candidateTime >= 0 {
																	fmt.Printf("[PARSER] DEBUG: Slot %d - candidateTime: %d, inWindow: true\n", tSlot, candidateTime)
																	if earliestFirstHoverTime == -1 || candidateTime < earliestFirstHoverTime {
																		earliestFirstHoverTime = candidateTime
																		earliestFirstHoverSlot = tSlot
																	}
																} else {
																	fmt.Printf("[PARSER] DEBUG: Slot %d - no valid hover time within window\n", tSlot)
																}
															}
														}
													} else {
														// When no overlapping slots, use last-hover algorithm for non-overlapping slots
														// Find the most recent valid hover time across slots that have recent hover sessions AND are within valid window
														if targetLastValidHoverTime[i] != nil {
															for tSlot, targetValidHoverTime := range targetLastValidHoverTime[i] {
																// Only consider non-overlapping slots that have recent hover sessions
																if tSlot != 7 && tSlot != 8 && tSlot != 9 {
																	if _, hasRecentSession := hoverDurations[i][tSlot]; hasRecentSession {
																		// Only consider last hover times that are within the valid window (relative to confirm box entry)
																		tickDiff := startTick - targetValidHoverTime
																		if tickDiff >= 0 && tickDiff <= validReportWindowSeconds*ticksPerSecond {
																			if targetValidHoverTime > mostRecentLastHoverTime {
																				mostRecentLastHoverTime = targetValidHoverTime
																				mostRecentLastHoverSlot = tSlot
																			}
																		}
																	}
																}
															}
														}
													}

													hasNonOverlappingSlots := mostRecentLastHoverTime >= 0
													if hasOverlappingSlots {
														fmt.Printf("[PARSER] DEBUG: Earliest first hover time: %d (Slot %d)\n", earliestFirstHoverTime, earliestFirstHoverSlot)
													}
													if hasNonOverlappingSlots {
														fmt.Printf("[PARSER] DEBUG: Most recent last hover time: %d (Slot %d)\n", mostRecentLastHoverTime, mostRecentLastHoverSlot)
													}
													// Only create report if hover duration was sufficient and there's been a valid hover session recently
													// Note: We don't pre-check validReference here - each candidate is validated individually
													// during scoring, and if no valid candidates exist, bestTarget will remain -1
													if confirmBoxHoverDuration >= minHoverDuration && (hasOverlappingSlots || hasNonOverlappingSlots) {

														fmt.Printf("[PARSER] ConfirmBoxHover Confirmed!")

														bestTarget := -1
														maxScore := 0.0

														type candidate struct {
															slot                 int
															duration             int
															targetTickDiff       int
															targetValidHoverTick int
															isLastHovered        int
															isFirstHovered       int
															score                float64
															name                 string
															hero                 string
														}
														var candidates []candidate

														var filteredOutCandidates []candidate
														for tSlot, duration := range hoverDurations[i] {
															if duration < minHoverDuration {
																candName := player_resources[tSlot].Name
																if candName == "" {
																	candName = fmt.Sprintf("Slot %d", tSlot)
																}
																candHero := player_resources[tSlot].Hero
																if candHero == "" {
																	candHero = "Unknown"
																}
																filteredOutCandidates = append(filteredOutCandidates, candidate{
																	slot:     tSlot,
																	duration: duration,
																	name:     candName,
																	hero:     candHero,
																})
																continue
															}
															if duration >= minHoverDuration {
																var targetValidHoverTime int
																var exists bool
																var isFirstHovered, isLastHovered int
																var score float64

																if hasOverlappingSlots {
																	// When overlapping slots are present, use first-hover algorithm for ALL slots
																	exists = false
																	// Try targetFirstValidHoverTime first
																	if targetFirstValidHoverTime[i] != nil {
																		if targetValidHoverTime, exists = targetFirstValidHoverTime[i][tSlot]; exists {
																			targetTickDiff := startTick - targetValidHoverTime
																			if targetTickDiff < 0 || targetTickDiff > validReportWindowSeconds*ticksPerSecond {
																				exists = false
																			}
																		}
																	}
																	// Fallback to targetLastValidHoverTime if first is missing or outside window
																	if !exists && targetLastValidHoverTime[i] != nil {
																		if targetValidHoverTime, exists = targetLastValidHoverTime[i][tSlot]; exists {
																			targetTickDiff := startTick - targetValidHoverTime
																			if targetTickDiff < 0 || targetTickDiff > validReportWindowSeconds*ticksPerSecond {
																				exists = false
																			}
																		}
																	}
																	if exists {
																		targetTickDiff := startTick - targetValidHoverTime
																		fmt.Printf("[PARSER] DEBUG: Scoring slot %d - targetValidHoverTime: %d, targetTickDiff: %d, inWindow: true\n", tSlot, targetValidHoverTime, targetTickDiff)
																		isFirstHovered = 0
																		if earliestFirstHoverTime >= 0 && targetValidHoverTime == earliestFirstHoverTime {
																			isFirstHovered = 1
																		}
																		score = float64(duration)*durationWeight + float64(isFirstHovered)*recencyWeight*5.0
																	} else {
																		fmt.Printf("[PARSER] DEBUG: Slot %d filtered out - no valid hover time within window\n", tSlot)
																		continue
																	}
																} else {
																	// When no overlapping slots, use last-hover algorithm for non-overlapping slots
																	// Check if this slot's last valid hover time matches the most recent across all slots
																	// Only one slot will match (the one with the highest targetLastValidHoverTime)
																	if targetValidHoverTime, exists = targetLastValidHoverTime[i][tSlot]; exists {
																		targetTickDiff := startTick - targetValidHoverTime
																		if targetTickDiff >= 0 && targetTickDiff <= validReportWindowSeconds*ticksPerSecond {
																			isLastHovered = 0
																			if mostRecentLastHoverTime >= 0 && targetValidHoverTime == mostRecentLastHoverTime {
																				isLastHovered = 1
																			}
																			score = float64(duration)*durationWeight + float64(isLastHovered)*recencyWeight*5.0
																		} else {
																			continue
																		}
																	} else {
																		continue
																	}
																}

																targetTickDiff := startTick - targetValidHoverTime
																candName := player_resources[tSlot].Name
																if candName == "" {
																	candName = fmt.Sprintf("Slot %d", tSlot)
																}
																candHero := player_resources[tSlot].Hero
																if candHero == "" {
																	candHero = "Unknown"
																}

																candidates = append(candidates, candidate{
																	slot:                 tSlot,
																	duration:             duration,
																	targetTickDiff:       targetTickDiff,
																	targetValidHoverTick: targetValidHoverTime,
																	isLastHovered:        isLastHovered,
																	isFirstHovered:       isFirstHovered,
																	score:                score,
																	name:                 candName,
																	hero:                 candHero,
																})

																if score > maxScore {
																	maxScore = score
																	bestTarget = tSlot
																}
															}
														}

														if bestTarget != -1 && bestTarget != i {
															reporterName := name
															if reporterName == "" {
																reporterName = fmt.Sprintf("Slot %d", i)
															}
															reporterHero := player_resources[i].Hero
															if reporterHero == "" {
																reporterHero = "Unknown"
															}

															fmt.Printf("[PARSER] Report detection (Time: %s) - Reporter: %s (%s, Slot %d)\n", timestamp, reporterName, reporterHero, i)
															fmt.Printf("[PARSER] Confirm box hover duration: %d ticks\n", confirmBoxHoverDuration)
															if hasOverlappingSlots {
																if earliestFirstHoverSlot >= 0 {
																	fmt.Printf("[PARSER] Using first-hover algorithm (slots 6/7/8 detected) - Reference: Slot %d at tick %d (earliest first hover)\n", earliestFirstHoverSlot, earliestFirstHoverTime)
																} else {
																	fmt.Printf("[PARSER] Using first-hover algorithm (slots 6/7/8 detected) - Reference: earliestFirstHoverTime = %d\n", earliestFirstHoverTime)
																}
															} else {
																if mostRecentLastHoverSlot >= 0 {
																	fmt.Printf("[PARSER] Using last-hover algorithm - Reference: Slot %d at tick %d (most recent hover)\n", mostRecentLastHoverSlot, mostRecentLastHoverTime)
																} else {
																	fmt.Printf("[PARSER] Using last-hover algorithm - Reference: mostRecentLastHoverTime = %d\n", mostRecentLastHoverTime)
																}
															}
															fmt.Printf("[PARSER] Weighting parameters - durationWeight: %.2f, recencyWeight: %.2f, recencyMultiplier: 5.0, minHoverDuration: %d ticks\n", durationWeight, recencyWeight, minHoverDuration)
															fmt.Printf("[PARSER] Score formula: score = (duration × %.2f) + (recencyBonus × %.2f × 5.0) where recencyBonus = 1 if hover matches reference, else 0\n", durationWeight, recencyWeight)

															if len(filteredOutCandidates) > 0 {
																fmt.Printf("[PARSER] Filtered out %d candidate(s) due to insufficient hover duration (< %d ticks):\n", len(filteredOutCandidates), minHoverDuration)
																for _, cand := range filteredOutCandidates {
																	fmt.Printf("[PARSER]   - Slot %d: %s (%s) - Duration: %d ticks (below minimum of %d)\n", cand.slot, cand.name, cand.hero, cand.duration, minHoverDuration)
																}
															}

															fmt.Printf("[PARSER] Found %d candidate(s) for report attribution:\n", len(candidates))

															for _, cand := range candidates {
																marker := " "
																if cand.slot == bestTarget {
																	marker = "★"
																}
																sessionCount := 0
																sessionDetails := ""
																if sessions, exists := hoverSessionDurations[i][cand.slot]; exists {
																	sessionCount = len(sessions)
																	if sessionCount > 0 {
																		sessionDetails = fmt.Sprintf(" (sessions: %v, max: %d)", sessions, cand.duration)
																	}
																}
																hoverType := "LastHovered"
																hoverValue := cand.isLastHovered
																recencyBonus := float64(cand.isLastHovered) * recencyWeight * 5.0
																if hasOverlappingSlots {
																	hoverType = "FirstHovered"
																	hoverValue = cand.isFirstHovered
																	recencyBonus = float64(cand.isFirstHovered) * recencyWeight * 5.0
																}
																durationScore := float64(cand.duration) * durationWeight
																fmt.Printf("[PARSER]   %s Slot %d: %s (%s) - Duration: %d ticks%s, HoverTick: %d, TickDiff: %d, %s: %d\n",
																	marker, cand.slot, cand.name, cand.hero, cand.duration, sessionDetails, cand.targetValidHoverTick, cand.targetTickDiff, hoverType, hoverValue)
																fmt.Printf("[PARSER]     Score breakdown: (%.0f × %.2f) + (%.0f × %.2f × 5.0) = %.2f + %.2f = %.2f\n",
																	float64(cand.duration), durationWeight, float64(hoverValue), recencyWeight, durationScore, recencyBonus, cand.score)
															}

															if len(candidates) > 1 {
																fmt.Printf("[PARSER] Selected Slot %d (Score: %.2f) as best target\n", bestTarget, maxScore)
															}

															finalTargetSlot := bestTarget

															if finalTargetSlot >= 0 && finalTargetSlot < 10 {
																targetSteamID := player_resources[finalTargetSlot].SteamID
																targetName := player_resources[finalTargetSlot].Name
																targetHero := player_resources[finalTargetSlot].Hero

																if targetSteamID == 0 {
																	if foundSteamID, exists := playerSteamIDs[finalTargetSlot]; exists && foundSteamID > 0 {
																		targetSteamID = foundSteamID
																	} else {
																		fmt.Printf("[PARSER] =================== End Report Detection (Reason: Target SteamID not found for slot %d) ===================\n\n", finalTargetSlot)
																		delete(confirmBoxHoverStart, i)
																		continue
																	}
																}

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

																sort.Slice(candidates, func(a, b int) bool {
																	return candidates[a].score > candidates[b].score
																})

																topCandidates := make([]Candidate, 0)
																maxCandidates := 3
																if len(candidates) < maxCandidates {
																	maxCandidates = len(candidates)
																}
																for idx := 0; idx < maxCandidates; idx++ {
																	cand := candidates[idx]
																	candSteamID := player_resources[cand.slot].SteamID
																	if candSteamID == 0 {
																		if foundSteamID, exists := playerSteamIDs[cand.slot]; exists && foundSteamID > 0 {
																			candSteamID = foundSteamID
																		}
																	}
																	topCandidates = append(topCandidates, Candidate{
																		Slot:    cand.slot,
																		SteamID: candSteamID,
																		Name:    cand.name,
																		Hero:    cand.hero,
																		Score:   cand.score,
																	})
																}

																newReport := &Report{
																	Time:                   fmt.Sprintf("%02d:%02d", minutes, secs),
																	SteamID:                steamid,
																	Slot:                   i,
																	Name:                   name,
																	Team:                   reportTeam,
																	Hero:                   player_resources[i].Hero,
																	TargetSlot:             finalTargetSlot,
																	TargetSteamID:          targetSteamID,
																	TargetName:             targetName,
																	TargetHero:             targetHero,
																	Candidates:             topCandidates,
																	SelectedCandidateIndex: 0,
																}

																fmt.Printf("[PARSER] Created report - Time: %s, Reporter: %s (Slot %d) -> Target: %s (Slot %d, SteamID: %d)\n",
																	newReport.Time, newReport.Name, newReport.Slot, newReport.TargetName, newReport.TargetSlot, newReport.TargetSteamID)
																fmt.Printf("[PARSER] =================== End Report Detection (Reason: Report created successfully) ===================\n\n")

																reports = append(reports, newReport)

																delete(hoverDurations, i)
																delete(lastHoverTime, i)
																delete(targetLastHoverTime, i)
																delete(targetLastValidHoverTime, i)
																delete(hoverSessionStart, i)
																delete(hoverSessionDurations, i)
															} else {
																fmt.Printf("[PARSER] =================== End Report Detection (Reason: Invalid target slot %d) ===================\n\n", finalTargetSlot)
															}
														}
													} else {
														var reason string
														if confirmBoxHoverDuration < minHoverDuration {
															reason = fmt.Sprintf("Insufficient hover duration (%d ticks < %d ticks minimum)", confirmBoxHoverDuration, minHoverDuration)
														} else if !hasOverlappingSlots && !hasNonOverlappingSlots {
															reason = "No valid hover slots detected"
														} else {
															reason = "Unknown reason"
														}
														fmt.Printf("[PARSER] =================== End Report Detection (Reason: %s) ===================\n\n", reason)
													}
													delete(confirmBoxHoverStart, i)
												}
											}

											if lastTick, exists := lastHoverTime[i]; exists {
												if current_tick-lastTick > validReportWindowSeconds*ticksPerSecond {
													// Finalize any active sessions before clearing
													for tSlot, sessionStart := range hoverSessionStart[i] {
														if sessionStart > 0 {
															sessionDuration := current_tick - sessionStart
															if sessionDuration > 0 {
																hoverSessionDurations[i][tSlot] = append(hoverSessionDurations[i][tSlot], sessionDuration)
																if sessionDuration > hoverDurations[i][tSlot] {
																	hoverDurations[i][tSlot] = sessionDuration
																}
																// Update valid hover time if session meets minimum duration
																if sessionDuration >= minHoverDuration {
																	if targetLastValidHoverTime[i] == nil {
																		targetLastValidHoverTime[i] = make(map[int]int)
																	}
																	targetLastValidHoverTime[i][tSlot] = current_tick
																}
															}
														}
													}
													delete(hoverDurations, i)
													delete(lastHoverTime, i)
													delete(targetLastHoverTime, i)
													delete(targetLastValidHoverTime, i)
													delete(hoverSessionStart, i)
													delete(hoverSessionDurations, i)
												}
											}
											break
										}
									}
								}
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

	for steamid, startTick := range scoreboardOpenStartTick {
		if begin_tick > 0 {
			startMinutes, startSecs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, startTick)
			endMinutes, endSecs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, current_tick)
			timeRange := TimeRange{
				Start: fmt.Sprintf("%02d:%02d", startMinutes, startSecs),
				End:   fmt.Sprintf("%02d:%02d", endMinutes, endSecs),
			}
			scoreboardTimeRanges[steamid] = append(scoreboardTimeRanges[steamid], timeRange)
		}
	}

	fmt.Printf("[PARSER] Scoreboard tracking summary - open ranges: %d, completed ranges: %d\n", len(scoreboardOpenStartTick), len(scoreboardTimeRanges))

	scoreboardActivities := make([]*ScoreboardActivity, 0)
	for i := 0; i < 10; i++ {
		if player_resources[i].SteamID > 0 {
			steamid := player_resources[i].SteamID
			if ranges, exists := scoreboardTimeRanges[steamid]; exists && len(ranges) > 0 {
				scoreboardActivities = append(scoreboardActivities, &ScoreboardActivity{
					SteamID:    steamid,
					Slot:       i,
					Name:       player_resources[i].Name,
					Hero:       player_resources[i].Hero,
					Team:       player_resources[i].Team,
					TimeRanges: ranges,
				})
			}
		}
	}
	fmt.Printf("[PARSER] Created %d scoreboard activity entries\n", len(scoreboardActivities))

	players := make([]PlayerResource, 10)
	copy(players[:], player_resources[:])

	return ParseResult{
		MatchID:              matchID,
		TeamReports:          teamReports,
		EnemyReports:         enemyReports,
		Reports:              reports,
		ScoreboardActivities: scoreboardActivities,
		Players:              players,
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
