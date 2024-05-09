// useful examples @ https://github.com/HighGroundVision/Mango/blob/master/web.go
// go run . 2> debug.log
// opendota : parse : replay url : http://replay273.valve.net/570/7697260946_320080104.dem.bz2
/*
CDOTA_PlayerResource.
"m_vecPlayerData.0001.m_iRankTier": (int32) 65,
Ranks are represented by integers:
  10-15: Herald
  20-25: Guardian
  30-35: Crusader
  40-45: Archon
  50-55: Legend
  60-65: Ancient
  70-75: Divine
  80-85: Immortal
  Each increment represents an additional star

  Assume 6 behavior levels and 4 comm levels

  "m_vecPlayerData.0006.m_nCommLevel": (int32) 6,
  "m_vecPlayerData.0006.m_nBehaviorLevel": (int32) 4,

*/
package main

import (
  "log"
  "os"
  "fmt"
  "time"
  "math"

  "github.com/dotabuff/manta"
  "github.com/dotabuff/manta/dota"
)

//7728014104 <- in this matchid i purposely moved cursor to corners of screen 1080p during scoreboard open to examine values.
var matchid string = "7697260946"
var replay_path string = fmt.Sprintf("/mnt/c/Program Files (x86)/Steam/steamapps/common/dota 2 beta/game/dota/replays/%s.dem", matchid)
var current_tick int = 0
var begin_tick int = 0
var scoreboardOpen map[uint64]bool //steamid as index

var reportedSteamID uint64 = 76561197971316129
var reportedTeam int = 2
var reportedSlot int = -1
var gameTime time.Duration
var pausedTicks int = 0

type PlayerResource struct {
    steamid uint64
    entindex  uint32
    team int32 //2 = radiant, 3 = dire
    name string
}

var player_resources [10]PlayerResource

var hasReportedYou [10]bool
var teamReports int = 0
var enemyReports int = 0

// // A message that has been read from an outerMessage but not yet processed.
// type pendingMessage struct {
// 	tick uint32
// 	t    int32
// 	buf  []byte
// }

// // Provides a sortable structure for storing messages in the same packet.
// type pendingMessages []*pendingMessage

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

/*
27:3
22:46

30*60*4 + 17*30 = 10290
*/
func ticksToMinutesAndSeconds(ticks int) (int, int) {
	// subtract 4 min 17 seconds
	//ticks = ticks - 7710
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

func printFields(data interface{}) {
  // Type assertion to check if data is a map[string]interface{}
  if values, ok := data.(map[string]interface{}); ok {
    // Iterate over the map and print key-value pairs
    for key, value := range values {
      fmt.Printf("Key: %s, Value: %v\n", key, value)
    }
  } else {
    fmt.Printf("Data is not a map[string]interface{}, cannot print fields\n")
  }
}


func isReportButton(x int, y int) int {
  if x >= 865 && x <= 893 {
    if y >= 100 && y <= 128 {
      return 0
    } else if y >= 156 && y <= 184 {
      return 1
    } else if y >= 214 && y <= 242 {
      return 2
    } else if y >= 271 && y <= 299 {
      return 3
    } else if y >= 328 && y <= 356 {
      return 4
    } else if y >= 418 && y <= 446 {
      return 5
    } else if y >= 475 && y <= 503 {
      return 6
    } else if y >= 532 && y <= 560 {
      return 7
    } else if y >= 589 && y <= 617 {
      return 8
    } else if y >= 646 && y <= 674 {
      return 9
    }
  }
  return -1
}

func main() {
  fmt.Printf("=========\n")
  fmt.Printf("=========Parse Init=========\n")
  fmt.Printf("=========\n")
  // Initialize the map
  scoreboardOpen = make(map[uint64]bool)

  // Create a new parser instance from a file. Alternatively see NewParser([]byte)
  f, err := os.Open(replay_path)
  if err != nil {
    log.Fatalf("unable to open file: %s", err)
  }
  defer f.Close()

  p, err := manta.NewStreamParser(f)
  if err != nil {
    log.Fatalf("unable to create parser: %s", err)
  }

/*
  p.Callbacks.OnCDemoSyncTick( func(m *dota.CDemoSyncTick) error {
  	fmt.Printf("Sync Tick")
  	return nil
  })
  */
  /*
  p.Callbacks.OnCDemoPacket( func (m *dota.CDemoPacket) error {

  	msg := m.GetData()
  	buffer := make([]byte, len(msg))
  	copy(buffer, msg)
  	r := &reader{buffer, uint32(len(msg)), 0, 0, 0}
  	t := int32(r.readUBitVar())
  	// fmt.Printf("%d ... ", t)

  	messageType := dota.EDotaUserMessages(t)
  	switch messageType {
  	  case dota.EDotaUserMessages_DOTA_UM_ChatMessage: //612
  	    fmt.Printf("Chat Packet")
      case dota.EDotaUserMessages_DOTA_UM_SpectatorPlayerClick:
        fmt.Printf("PlayerClick Packet")
  	}

 	// messageTypeGC := dota.EDOTAGCMsg(t)
  	// switch messageTypeGC {
  	// 	case dota.EDOTAGCMsg_k_EMsgGCSubmitPlayerReport:
  	// 		fmt.Printf("GC Packet")
  	// 	case dota.EDOTAGCMsg_k_EMsgGCSubmitPlayerReportV2:
  	// 		fmt.Printf("GC Packet2")
  	// }
  	return nil
  })
  */

  //Dota 2 = 30 ticks/sec.
  p.Callbacks.OnCNETMsg_Tick( func(m * dota.CNETMsg_Tick) error {
  	//fmt.Printf("Tick ... %d %d",m.GetTick())
  	current_tick = int(m.GetTick())
  	return nil
  })
  /*
  	"SIGNONSTATE_NONE":        0,
  	"SIGNONSTATE_CHALLENGE":   1,
  	"SIGNONSTATE_CONNECTED":   2,
  	"SIGNONSTATE_NEW":         3,
  	"SIGNONSTATE_PRESPAWN":    4,
  	"SIGNONSTATE_SPAWN":       5,
  	"SIGNONSTATE_FULL":        6,
  	"SIGNONSTATE_CHANGELEVEL": 7,
  */
  /*
  p.Callbacks.OnCNETMsg_SignonState( func(m *dota.CNETMsg_SignonState) error {
  	fmt.Printf("State changed to : %d",int(m.GetSignonState()))
  	return nil
  })
  */

    /*
  DOTA_GAMERULES_STATE_INIT 0 
  DOTA_GAMERULES_STATE_WAIT_FOR_PLAYERS_TO_LOAD 1 
  DOTA_GAMERULES_STATE_CUSTOM_GAME_SETUP  2 
  DOTA_GAMERULES_STATE_HERO_SELECTION 3 
  DOTA_GAMERULES_STATE_STRATEGY_TIME  4 
  DOTA_GAMERULES_STATE_TEAM_SHOWCASE  5 
  DOTA_GAMERULES_STATE_PRE_GAME 6 
  DOTA_GAMERULES_STATE_GAME_IN_PROGRESS 7 
  DOTA_GAMERULES_STATE_POST_GAME  8 
  DOTA_GAMERULES_STATE_DISCONNECT 9
  9,2,12,3,8,10,4,5

  6,7
  */
  p.Callbacks.OnCDOTAUserMsg_GamerulesStateChanged(func(m *dota.CDOTAUserMsg_GamerulesStateChanged) error {
    if m.GetState() == 5 {
      // Game Begins. 00:00
      begin_tick = current_tick
    }
    // minutes,secs := ticksToMinutesAndSeconds(current_tick)
    // fmt.Printf("Game state is now : %d at (%d,%d)",m.GetState(),minutes,secs)
    return nil
  })

  /*
  // Register a callback, this time for the OnCUserMessageSayText2 event.
  p.Callbacks.OnCUserMessageSayText2(func(m *dota.CUserMessageSayText2) error {
    fmt.Printf("OnCUserMessageSayText2 %s said: %s\n", m.GetParam1(), m.GetParam2())
    return nil
  })

    // Register a callback, this time for the OnCUserMessageSayText2 event.
  p.Callbacks.OnCUserMessageSayText(func(m *dota.CUserMessageSayText) error {
    fmt.Printf("OnCUserMessageSayText %s said: %s\n", m.GetPlayerindex(), m.GetText())
    return nil
  })

  */
  // Register a callback, this time for the OnCUserMessageSayText2 event.
  p.Callbacks.OnCUserMessageSayTextChannel(func(m *dota.CUserMessageSayTextChannel) error {
    fmt.Printf("CUserMessageSayTextChannel %s said: %s\n", player_resources[m.GetPlayer()].name, m.GetText())
    return nil
  })

  //CDOTAUserMsg_ChatEvent <- This one is for all Update Messages like Barracks destroyed, Roshan killed etc.

  //PlayerId from 1-10 here.
  p.Callbacks.OnCDOTAUserMsg_ChatMessage(func(m *dota.CDOTAUserMsg_ChatMessage) error {
    team := ""
    if (m.GetSourcePlayerId() > 5 && reportedTeam > 5) || (m.GetSourcePlayerId() < 6 && reportedTeam < 6) {
      team = "FRIENDLY"
    } else {
      team = "ENEMY"
    }
    team2 := ""
    if m.GetSourcePlayerId() > 5 {
      team2 = "dire"
    } else {
      team2 = "radiant"
    }
  	 
  	minutes,secs := ticksToMinutesAndSeconds(current_tick)
  	fmt.Printf("_ALLCHAT_ (%02d mins:%02d seconds) [ (%s_%s) %s ] said: %s\n",minutes,secs,team,team2, player_resources[m.GetSourcePlayerId()].name, m.GetMessageText())

  	return nil
  })

  //Doesn't fire
  /*
  p.Callbacks.OnCDOTAUserMsg_SpectatorPlayerClick(func(m *dota.CDOTAUserMsg_SpectatorPlayerClick) error {
    fmt.Printf("Player Click Order:%d Ent:%d", m.GetOrderType(),m.GetEntindex())
    return nil
  })
  */

  /*
  Entindex       *int32      `protobuf:"varint,1,opt,name=entindex" json:"entindex,omitempty"`
  OrderType      *int32      `protobuf:"varint,2,opt,name=order_type,json=orderType" json:"order_type,omitempty"`
  Units          []int32     `protobuf:"varint,3,rep,name=units" json:"units,omitempty"`
  TargetIndex    *int32      `protobuf:"varint,4,opt,name=target_index,json=targetIndex" json:"target_index,omitempty"`
  AbilityId      *int32      `protobuf:"varint,5,opt,name=ability_id,json=abilityId" json:"ability_id,omitempty"`
  Position       *CMsgVector `protobuf:"bytes,6,opt,name=position" json:"position,omitempty"`
  Queue          *bool       `protobuf:"varint,7,opt,name=queue" json:"queue,omitempty"`
  SequenceNumber *int32      `protobuf:"varint,8,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
  Flags          *uint32     `protobuf:"varint,9,opt,name=flags" json:"flags,omitempty"`
  */
  /*
  p.Callbacks.OnCDOTAUserMsg_SpectatorPlayerUnitOrders(func(m *dota.CDOTAUserMsg_SpectatorPlayerUnitOrders) error {
    //fmt.Printf("Player Unit Order... %d", m.GetEntindex())
    orderingPlayer := -1
    for i := 0; i < 10; i++ {
      if int(player_resources[i].entindex) == int(m.GetEntindex()) {
        orderingPlayer = i
        break
      }
    }
    if orderingPlayer != -1 {
      //fmt.Printf("OrderingPlayer steam id == %d",player_resources[orderingPlayer].steamid)
      //fmt.Printf("report steam id == %d",reportedSteamID)
      if player_resources[orderingPlayer].steamid == reportedSteamID {
        //minutes,secs := ticksToMinutesAndSeconds(current_tick)
        //fmt.Printf("Reported player issued order at {%d:%d}",minutes,secs)
      }
    }
    return nil
  })
  */
  /*
  p.onCSVCMsg_PacketEntities(func(pe *manta.PacketEntity, pet manta.EntityEventType) error {

    if pe.ClassName == "C_DOTAPlayer" {

    }
    return nil
  })
  */
  //look at manta_test.go for example.
  //EntityOp == int
  /*
    e.Dump() ==
      // String returns a human identifiable string for the Entity
      func (e *Entity) String() string {
        return fmt.Sprintf("%d <%s>", e.index, e.class.name)
      }
      // Map returns a map of current entity state as key-value pairs
      func (e *Entity) Map() map[string]interface{} {
        values := make(map[string]interface{})
        for _, fp := range e.class.getFieldPaths(newFieldPath(), e.state) {
          values[e.class.getNameForFieldPath(fp)] = e.state.get(fp)
        }
        return values
      }
  */
  /*
    because the struct field does not start with Caps, its not public... ?? really??

    the m_nPlayerID is not translatable to slot, different in each game.
    Don't really need it though, I guess we filter by being on same team as the target steamID, yet not matching steamID.
  */
  p.OnEntity(func(e *manta.Entity, op manta.EntityOp) error {
    //e.Dump()
    //reportedSteamID

    if e.GetClassName() == "CDOTA_PlayerResource" {
      for i := 0; i < 10; i++ {
        isVictim := false
        if steamid,steamidok := e.GetUint64(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerSteamID",i)); steamidok {
          //fmt.Printf("Steam id is %d",steamid)
          player_resources[i].steamid = steamid

          if steamid == reportedSteamID {
            isVictim = true
            reportedSlot = i
          }
        }

        if entindex,entindexok := e.GetUint32(fmt.Sprintf("m_vecPlayerData.000%d.m_nPlayerSlot",i)); entindexok {
          //fmt.Printf("Entindex is %d",entindex)
          player_resources[i].entindex = entindex
        }

        if team,teamok := e.GetInt32(fmt.Sprintf("m_vecPlayerData.000%d.m_iPlayerTeam",i)); teamok {
          //fmt.Printf("Team is %d", team)
          player_resources[i].team = team
          if isVictim {
            reportedTeam = int(team)
          }
        }

        if name,nameok := e.GetString(fmt.Sprintf("m_vecPlayerData.000%d.m_iszPlayerName",i)); nameok {
          player_resources[i].name = name
        }
      }
    } else if e.GetClassName() == "CDOTAGamerulesProxy" {
      
      if v, ok := e.GetInt32("m_pGameRules.m_nTotalPausedTicks"); ok {
        if v > 0 {
          pausedTicks = int(v)
          //fmt.Printf("m_nTotalPausedTicks %d",v)
        }
      }
      
    }

    //Don't process before heroes picked. (game truely started)
    if begin_tick == 0 {
      return nil
    }
    if e.GetClassName() == "CDOTAPlayerController" {

      //m_nPlayerID
      if steamid,ok2 :=e.GetUint64("m_steamID"); ok2 {
        if name,ok3 := e.GetString("m_iszPlayerName");ok3 {
          if steamid != reportedSteamID {

            //fmt.Printf("steamid is == %d",steamid)
            if statsPanel, ok := e.GetInt32("m_iStatsPanel"); ok {
              if statsPanel != 0 && statsPanel != 1 {
                fmt.Printf("Detected statsPanel value other than 1 or 0 %d\n",steamid)
              }

              if statsPanel == 1 {
                  //activated from off state.
                  if !scoreboardOpen[steamid] {
                    // minutes,secs := ticksToMinutesAndSeconds(current_tick)
                    //fmt.Printf("Scoreboard open at time : {%d,%d}, player: %d , %s", minutes,secs,steamid,name)
                  }
                  //print mouse coords
                  if xpos,xposok := e.GetInt32("m_iCursor.0000"); xposok {
                    if ypos,yposok := e.GetInt32("m_iCursor.0001"); yposok {
                      xpos = int32(math.Round(float64(xpos)/510 * 1920))
                      ypos = int32(math.Round(float64(ypos)/383 * 1080))
                      //fmt.Printf("Mouse XPOS : %d, Mouse YPOS : %d",xpos,ypos)  
                      if targetSlot := isReportButton(int(xpos),int(ypos)); targetSlot != -1 {
                        if targetSlot == reportedSlot {
                          for i := 0; i < 10; i++ {
                            if player_resources[i].steamid == steamid {
                              
                              //found the reporting player's slot id.
                              
                              if !hasReportedYou[i] {
                                minutes,secs := ticksToMinutesAndSeconds(current_tick)
                                if team,okteam := e.GetUint64("m_iTeamNum");okteam {
                                  if reportedTeam == int(team) {
                                    fmt.Printf("_REPORT_ (%02d mins:%02d seconds) from _TEAMMATE_: steamid=%d , slot=%d, name=%s\n",minutes,secs,steamid,i,name)
                                    teamReports += 1
                                  } else {
                                    fmt.Printf("_REPORT_ (%02d mins:%02d seconds) from _ENEMY_: steamid=%d , slot=%d, name=%s\n",minutes,secs,steamid,i,name)
                                    enemyReports += 1
                                  }
                                }
                              }
                              hasReportedYou[i] = true

                              //fmt.Printf("Who: %d Mouse XPOS : %d, Mouse YPOS : %d",i,xpos,ypos)
                              break
                            }
                          } 
                        }
                      }
                    }
                  }
                  scoreboardOpen[steamid] = true
              } else if statsPanel == 0 {

                //closed from on state.
                if scoreboardOpen[steamid] {
                  // minutes,secs := ticksToMinutesAndSeconds(current_tick)
                  //fmt.Printf("Scoreboard closed at time : {%d:%d}, player: %d , %s", minutes,secs,steamid,name )
                }
                
                scoreboardOpen[steamid] = false
              }
              
            }//statspanel
          } //mysteamid
        } //name
      } else {
        fmt.Printf("Cant get player id\n")
      }

    } 

    return nil
  })

  //https://github.com/dotabuff/manta/issues/73
  //DOTA_UM_SpectatorPlayerClick
  //C_DOTAPlayer

  // p.Callbacks.OnCMsgDOTASubmitPlayerReport(func(m *dota.CMsgDOTASubmitPlayerReport) error {
  // 	fmt.Printf("SOMEONE USED REPORT")
  // 	return nil
  // })

  // p.Callbacks.onCMsgDOTASubmitPlayerReport = append(p.Callbacks.onCMsgDOTASubmitPlayerReport, func(m *dota.CMsgDOTASubmitPlayerReport) error {
  // 	fmt.Printf("SOMEONE USED REPORT")
  // 	return nil
  // })

  // Start parsing the replay!
  p.Start()
  fmt.Printf("=========\n")
  fmt.Printf("=========Parse Complete=========\n")
  fmt.Printf("=========\n")
  fmt.Printf("%d reports from own team\n",teamReports)
  fmt.Printf("%d reports from enemy team\n",enemyReports)
}