package main

import (
  "log"
  "os"

  "github.com/dotabuff/manta"
  "github.com/dotabuff/manta/dota"
)
//7719413125
//7724919452
var replay_path string = "/mnt/c/Program Files (x86)/Steam/steamapps/common/dota 2 beta/game/dota/replays/7724919452.dem"
var current_tick int = 0
var begin_tick int = 0

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
	ticks = ticks - begin_tick
    totalSeconds := ticks / 30
    minutes := totalSeconds / 60
    seconds := totalSeconds % 60
    return minutes, seconds
}

func main() {
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
  	log.Printf("Sync Tick")
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
  	// log.Printf("%d ... ", t)

  	messageType := dota.EDotaUserMessages(t)
  	switch messageType {
  	case dota.EDotaUserMessages_DOTA_UM_ChatMessage: //612
  		log.Printf("Chat Packet")
  	}

 	// messageTypeGC := dota.EDOTAGCMsg(t)
  	// switch messageTypeGC {
  	// 	case dota.EDOTAGCMsg_k_EMsgGCSubmitPlayerReport:
  	// 		log.Printf("GC Packet")
  	// 	case dota.EDOTAGCMsg_k_EMsgGCSubmitPlayerReportV2:
  	// 		log.Printf("GC Packet2")
  	// }
  	return nil
  })
*/
  //Dota 2 = 30 ticks/sec.
  p.Callbacks.OnCNETMsg_Tick( func(m * dota.CNETMsg_Tick) error {
  	//log.Printf("Tick ... %d %d",m.GetTick())
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
  p.Callbacks.OnCNETMsg_SignonState( func(m *dota.CNETMsg_SignonState) error {
  	log.Printf("State changed to : %d",int(m.GetSignonState()))
  	return nil
  })

  // Register a callback, this time for the OnCUserMessageSayText2 event.
  p.Callbacks.OnCUserMessageSayText2(func(m *dota.CUserMessageSayText2) error {
    log.Printf("OnCUserMessageSayText2 %s said: %s\n", m.GetParam1(), m.GetParam2())
    return nil
  })

    // Register a callback, this time for the OnCUserMessageSayText2 event.
  p.Callbacks.OnCUserMessageSayText(func(m *dota.CUserMessageSayText) error {
    log.Printf("OnCUserMessageSayText %s said: %s\n", m.GetPlayerindex(), m.GetText())
    return nil
  })

    // Register a callback, this time for the OnCUserMessageSayText2 event.
  p.Callbacks.OnCUserMessageSayTextChannel(func(m *dota.CUserMessageSayTextChannel) error {
    log.Printf("CUserMessageSayTextChannel %s said: %s\n", m.GetPlayer(), m.GetText())
    return nil
  })

  //CDOTAUserMsg_ChatEvent - what is this one?
  p.Callbacks.OnCDOTAUserMsg_ChatMessage(func(m *dota.CDOTAUserMsg_ChatMessage) error {
  	team := "RADIANT"
  	if m.GetSourcePlayerId() > 5 {
  		team = "DIRE"
  	}
  	 
  	minutes,secs := ticksToMinutesAndSeconds(current_tick)
  	log.Printf("CDOTAUserMsg_ChatMessage (%d:%d) [%s-%d] said: %s\n",minutes,secs,team, m.GetSourcePlayerId(), m.GetMessageText())

  	return nil
  })
  /*
  DOTA_GAMERULES_STATE_INIT	0	
  DOTA_GAMERULES_STATE_WAIT_FOR_PLAYERS_TO_LOAD	1	
  DOTA_GAMERULES_STATE_CUSTOM_GAME_SETUP	2	
  DOTA_GAMERULES_STATE_HERO_SELECTION	3	
  DOTA_GAMERULES_STATE_STRATEGY_TIME	4	
  DOTA_GAMERULES_STATE_TEAM_SHOWCASE	5	
  DOTA_GAMERULES_STATE_PRE_GAME	6	
  DOTA_GAMERULES_STATE_GAME_IN_PROGRESS	7	
  DOTA_GAMERULES_STATE_POST_GAME	8	
  DOTA_GAMERULES_STATE_DISCONNECT	9
  9,2,12,3,8,10,4,5

  6,7
  */
  p.Callbacks.OnCDOTAUserMsg_GamerulesStateChanged(func(m *dota.CDOTAUserMsg_GamerulesStateChanged) error {

  	if m.GetState() == 5 {
  		begin_tick = current_tick
  	}
  	// minutes,secs := ticksToMinutesAndSeconds(current_tick)
  	// log.Printf("Game state is now : %d at (%d,%d)",m.GetState(),minutes,secs)
  	return nil
  })

  // p.Callbacks.OnCMsgDOTASubmitPlayerReport(func(m *dota.CMsgDOTASubmitPlayerReport) error {
  // 	log.Printf("SOMEONE USED REPORT")
  // 	return nil
  // })

  // p.Callbacks.onCMsgDOTASubmitPlayerReport = append(p.Callbacks.onCMsgDOTASubmitPlayerReport, func(m *dota.CMsgDOTASubmitPlayerReport) error {
  // 	log.Printf("SOMEONE USED REPORT")
  // 	return nil
  // })


  
  

  // Start parsing the replay!
  p.Start()

  log.Printf("Parse Complete!\n")
}