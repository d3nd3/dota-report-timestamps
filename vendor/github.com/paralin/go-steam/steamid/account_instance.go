package steamid

// DesktopInstance is the account instance value for a desktop.
const DesktopInstance uint32 = 1

// ConsoleInstance is the account instance value for a console.
const ConsoleInstance uint32 = 2

// WebInstance is the account instance value for mobile or web-based.
const WebInstance uint32 = 4

// AccountInstanceMask is used for packing chat instance flags in a steam ID.
const AccountInstanceMask uint32 = 0x000FFFFF

// ChatInstanceFlag is a flag a chat steam ID may have.
type ChatInstanceFlag uint32

const (
	// ChatInstanceFlagClan is set for clan based chat steam ids.
	ChatInstanceFlagClan ChatInstanceFlag = ChatInstanceFlag((AccountInstanceMask + 1) >> 1)
	// ChatInstanceFlagLobby is set for lobby based chat steam ids.
	ChatInstanceFlagLobby = ChatInstanceFlag((AccountInstanceMask + 1) >> 2)
	// ChatInstanceFlagMMSLobby is set for matchmaking lobby based chat steam ids.
	ChatInstanceFlagMMSLobby = ChatInstanceFlag((AccountInstanceMask + 1) >> 3)
)

// AccountInstance is an instance of an account.
type AccountInstance uint32

// HasFlag sees if the flag is set.
func (i AccountInstance) HasFlag(flag uint32) bool {
	return uint32(i)&flag != 0
}
