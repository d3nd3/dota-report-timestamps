# Setup
install golang  
git clone github.com/d3nd3/dota-report-timestamps  
cd dota-report-timestamps  
go mod download  

# Edit for your account
Edit main.go:  
var matchid string = "7697260946"  
var replay_dir string = "path/to/replays/"  

replace 7697260946 with the matchid you got reported in.  
replace the long replay_dir with your dota replays directory.  


# Running the replay parser
Ensure the replay is downloaded, download it from within Dota 2 Client or opendota website.  
`go run .`  
or optionally build it with `go build` then run the binary `./dota-report-timestamps`  

When ran without any argument it will display the players in the match, with thier slots and steam ids.  

Use that data to specify in commandline who is the victim you are interseted got reported.  
eg for see who reported blue slot 0 :  

```
go run . -s 0
```

eg for see who reported steam id 99999999 :

```
go run . -sid 99999999
```


# Troubleshooting
go get -u  
go mod tidy  

# Caveats
Only detects 16:9 ratio screens at the moment.  
Need help getting co-ordinate date for the other ratios.  

# Keywords
parse detect script detection behaviour reports report dota2 dota replay score false

