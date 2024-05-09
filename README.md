# Setup
install golang \ 
git clone github.com/d3nd3/dota-report-timestamps \ 
cd dota-report-timestamps \ 
go mod download\ 

# Edit for your account
Edit main.go: \ 
var matchid string = "7697260946" \
var replay_dir string = "path/to/replays/" \ 
var reportedSteamID uint64 = 76561197971316129 \ 

replace 7697260946 with the matchid you got reported in. \ 
replace the long replay_dir with your dota replays directory. \ 
replace `reportedSteamID` with your steamID. \ 

# Running the replay parser
Ensure the replay is downloaded, download it from within Dota 2 Client or opendota website. \ 
`go run .` \ 
or optionally build it with `go build` then run the binary `./dota-report-timestamps`

