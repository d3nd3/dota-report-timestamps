package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/dotabuff/manta"
	"github.com/dotabuff/manta/dota"
	"github.com/golang/protobuf/proto"
	"github.com/klauspost/compress/snappy"
)

type PlayerResource struct {
	SteamID  uint64
	EntIndex uint32
	Team     int32 // 2 = radiant, 3 = dire
	Name     string
}

type Report struct {
	Time                string
	Team                string // "FRIENDLY" or "ENEMY"
	SteamID             uint64
	Slot                int
	Name                string
	Confirmed           bool
	ConfirmationDelayMs int
}

type ParseResult struct {
	MatchID               int64     `json:"MatchID"`
	TeamReports           int       `json:"TeamReports"`
	EnemyReports          int       `json:"EnemyReports"`
	ConfirmedTeamReports  int       `json:"ConfirmedTeamReports"`
	ConfirmedEnemyReports int       `json:"ConfirmedEnemyReports"`
	Reports               []*Report `json:"Reports"`
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

func ParseReplay(matchID int64, file io.Reader, reportedSlot int, reportedSteamID uint64) (ParseResult, error) {
	var player_resources [10]PlayerResource

	var hasReportedYou [10]bool
	var teamReports int = 0
	var enemyReports int = 0
	var confirmedTeamReports int = 0
	var confirmedEnemyReports int = 0

	var current_tick int = 0
	var begin_tick int = 0
	var scoreboardOpen map[uint64]bool //steamid as index

	var reportedTeam int = 2
	var pausedTicks int = 0

	var reports []*Report

	// Track confirmation status: key is reporter slot (0-9)
	// Value is the tick when they last hovered the report button (if unconfirmed)
	lastHoverTick := make(map[int]int)

	// Map to access the Report object for each reporter
	// We assume one report attempt per reporter for simplicity/storage,
	// but allow updating it to 'Confirmed' if a confirmation follows.
	playerReportMap := make(map[int]*Report)

	// Use local variables that can be modified (shadow the parameters)
	// These need to be modifiable within closures
	actualReportedSlot := reportedSlot
	actualReportedSteamID := reportedSteamID

	// If SteamID is provided, reset slot to -1 to force lookup by SteamID
	// This prevents accidental default to slot 0 if reportedSlot was passed as 0 (default int)
	if actualReportedSteamID > 0 {
		actualReportedSlot = -1
	}

	// If only SteamID is provided, we'll find the slot later
	// If only slot is provided, we'll find the SteamID later

	// Initialize the map
	scoreboardOpen = make(map[uint64]bool)

	p, err := manta.NewStreamParser(file)
	if err != nil {
		return ParseResult{}, fmt.Errorf("unable to create parser: %s", err)
	}

	p.Callbacks.OnCNETMsg_Tick(func(m *dota.CNETMsg_Tick) error {
		current_tick = int(m.GetTick())
		return nil
	})

	p.Callbacks.OnCDOTAUserMsg_GamerulesStateChanged(func(m *dota.CDOTAUserMsg_GamerulesStateChanged) error {
		if m.GetState() == 5 {
			begin_tick = current_tick
		}
		return nil
	})

	p.OnEntity(func(e *manta.Entity, op manta.EntityOp) error {
		if e.GetClassName() == "CDOTA_PlayerResource" {
			for i := 0; i < 10; i++ {
				isVictim := false
				if actualReportedSlot != -1 && i == actualReportedSlot {
					isVictim = true
				}

				if steamid, steamidok := e.GetUint64(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerSteamID", i)); steamidok {
					player_resources[i].SteamID = steamid

					// If we're looking for a specific SteamID, find its slot
					if actualReportedSteamID > 0 {
						if steamid == actualReportedSteamID {
							isVictim = true
							actualReportedSlot = i
						}
					} else if isVictim {
						// If we're looking for a specific slot, get its SteamID
						actualReportedSteamID = steamid
					}
				}

				if entindex, entindexok := e.GetUint32(fmt.Sprintf("m_vecPlayerData.000%d.m_nPlayerSlot", i)); entindexok {
					player_resources[i].EntIndex = entindex
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
			}
			// gatheredResources = true

		} else if e.GetClassName() == "CDOTAGamerulesProxy" {

			if v, ok := e.GetInt32("m_pGameRules.m_nTotalPausedTicks"); ok {
				if v > 0 {
					pausedTicks = int(v)
				}
			}

		}
		if begin_tick == 0 {
			return nil
		}
		if e.GetClassName() == "CDOTAPlayerController" {

			if steamid, ok2 := e.GetUint64("m_steamID"); ok2 {

				if name, ok3 := e.GetString("m_iszPlayerName"); ok3 {
					if steamid != actualReportedSteamID {

						if statsPanel, ok := e.GetInt32("m_iStatsPanel"); ok {

							if statsPanel == 1 {
								//activated from off state.
								if !scoreboardOpen[steamid] {
								}
								//print mouse coords
								if xpos, xposok := e.GetInt32("m_iCursor.0000"); xposok {
									if ypos, yposok := e.GetInt32("m_iCursor.0001"); yposok {
										xpos = int32(math.Round(float64(xpos) / 510 * 1920))
										ypos = int32(math.Round(float64(ypos) / 383 * 1080))

										if aspect, aspectok := e.GetFloat32("m_flAspectRatio"); aspectok {
											targetSlot := isReportButton(int(xpos), int(ypos), aspect)

											// Check for confirmation (Cursor enters confirmation box)
											// Confirmation Box: X: [956, 1170], Y: [847, 888] (1920x1080 space)
											// We check this for ANY player who has an unconfirmed report pending
											for i := 0; i < 10; i++ {
												if player_resources[i].SteamID == steamid {
													// Found the reporter (i)

													// Check if this reporter has a pending report
													if report, exists := playerReportMap[i]; exists && !report.Confirmed {
														// Check if in confirmation box
														// Note: These coordinates are hardcoded based on user input for 1080p
														// xpos and ypos are already normalized to 1920x1080
														inConfirmBox := xpos >= 956 && xpos <= 1170 && ypos >= 847 && ypos <= 888

														if inConfirmBox {
															// Check time window (4 seconds = 120 ticks)
															tickDiff := current_tick - lastHoverTick[i]
															if tickDiff >= 0 && tickDiff <= 120 {
																report.Confirmed = true
																// tick diff * (1000ms / 30 ticks)
																report.ConfirmationDelayMs = int(float64(tickDiff) * (1000.0 / 30.0))

																if report.Team == "FRIENDLY" {
																	confirmedTeamReports++
																} else {
																	confirmedEnemyReports++
																}
															}
														}
													}

													if targetSlot != -1 {
														//It hovered over our report button.
														// Match original logic: check if targetSlot matches reportedSlot
														if targetSlot == actualReportedSlot {
															// Record the hover tick for confirmation window
															lastHoverTick[i] = current_tick

															// Create report if it doesn't exist
															if !hasReportedYou[i] {
																minutes, secs := ticksToMinutesAndSeconds(begin_tick, pausedTicks, current_tick)

																var reportTeam string
																if team, okteam := e.GetUint64("m_iTeamNum"); okteam {
																	reporterTeam := int(team)
																	if reportedTeam == reporterTeam {
																		reportTeam = "FRIENDLY"
																		teamReports += 1
																	} else {
																		reportTeam = "ENEMY"
																		enemyReports += 1
																	}
																}

																newReport := &Report{
																	Time:      fmt.Sprintf("%02d:%02d", minutes, secs),
																	SteamID:   steamid,
																	Slot:      i,
																	Name:      name,
																	Team:      reportTeam,
																	Confirmed: false,
																}

																reports = append(reports, newReport)
																playerReportMap[i] = newReport
																hasReportedYou[i] = true
															}
														}
													}
													break // Found reporter
												}
											}
										}
									}
								}
								scoreboardOpen[steamid] = true
							} else if statsPanel == 0 {
								scoreboardOpen[steamid] = false
							}

						} //statspanel
					} //mysteamid
				} //name
			} else {
				// fmt.Printf("Cant get player id\n")
			}

		}

		return nil
	})

	p.Start()

	return ParseResult{
		MatchID:               matchID,
		TeamReports:           teamReports,
		EnemyReports:          enemyReports,
		ConfirmedTeamReports:  confirmedTeamReports,
		ConfirmedEnemyReports: confirmedEnemyReports,
		Reports:               reports,
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
