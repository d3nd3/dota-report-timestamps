# Setup
install golang  
git clone github.com/d3nd3/dota-report-timestamps  
cd dota-report-timestamps  
go mod download  

# Setup Replay Directory.
Edit main.go:  
var replay_dir string = "path/to/replays/"  


# Running the replay parser
Ensure the replay is downloaded, download it from within Dota 2 Client or opendota website.  
`go run .`  
or optionally build it with `go build` then run the binary `./dota-report-timestamps`  

When ran without any argument it will display the players in the match, with thier slots and steam ids.  

Use that data to specify in commandline who is the victim you are interseted got reported.  
eg for see who reported blue slot 0 :  

```
go run . -s 0 -m 5430584395
```

eg for see who reported steam id 99999999 :

```
go run . -sid 99999999 -m 54356546456
```


# Troubleshooting
go get -u  
go mod tidy  

# Implementation Considerations
Aspect Ratio is obtained in the replay for each player  
When a player runs out of tips, the report button location shifts 100 pixels to the left  
When a player who has special hero like WK, he can have extra stuff in his scoreboard which makes his report button be in a different location.


# Keywords
parse detect script detection behaviour reports report dota2 dota replay score false

