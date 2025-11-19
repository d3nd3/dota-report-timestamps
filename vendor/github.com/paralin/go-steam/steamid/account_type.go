package steamid

import (
	"github.com/paralin/go-steam/protocol/steamlang"
)

// accountTypeChars maps account type to steam3 id chars.
var accountTypeChars = map[steamlang.EAccountType]rune{
	steamlang.EAccountType_AnonGameServer: 'A',
	steamlang.EAccountType_GameServer:     'G',
	steamlang.EAccountType_Multiseat:      'M',
	steamlang.EAccountType_Pending:        'P',
	steamlang.EAccountType_ContentServer:  'C',
	steamlang.EAccountType_Clan:           'g',
	steamlang.EAccountType_Chat:           'T',
	steamlang.EAccountType_Invalid:        'I',
	steamlang.EAccountType_Individual:     'U',
	steamlang.EAccountType_AnonUser:       'a',
}
