package steam

import (
	"math/rand"
	"time"

	"github.com/paralin/go-steam/netutil"
)

// CMServers contains a list of worldwide servers
//
// curl -s "https://api.steampowered.com/ISteamDirectory/GetCMListForConnect/v1/?key=$STEAM_API_KEY&format=json&celllid=1&cmtype=netfilter" | jq -r '.response.serverlist[] | .endpoint' | awk 'BEGIN {print "var CMServers = []string{"} {print "  \""$1"\","} END {print "}"}'
var CMServers = []string{
	"155.133.253.34:27017",
	"155.133.253.50:27017",
	"162.254.192.71:27017",
	"162.254.192.74:27017",
	"162.254.192.75:27017",
	"162.254.192.87:27017",
	"162.254.193.102:27017",
	"162.254.193.74:27017",
	"162.254.195.66:27017",
	"162.254.195.71:27017",
	"162.254.196.67:27017",
	"162.254.196.68:27017",
	"162.254.196.83:27017",
	"162.254.196.84:27017",
	"162.254.199.163:27017",
	"162.254.199.181:27017",
	"185.25.182.20:27017",
	"185.25.182.52:27017",
	"205.196.6.214:27017",
	"205.196.6.215:27017",
}

// GetRandomCM returns back a random server anywhere
func GetRandomCM() *netutil.PortAddr {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	servers := CMServers
	addr := netutil.ParsePortAddr(servers[rng.Int31n(int32(len(servers)))])
	if addr == nil {
		panic("invalid address in CMServers slice")
	}
	return addr
}
