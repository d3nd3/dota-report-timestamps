# Dota 2 Report Timestamp Tool

A tool to parse Dota 2 replay files (`.dem`) and detect when players opened the scoreboard and clicked the report button. This helps identify who reported whom and at what timestamp.

## Setup

1.  **Install Go**: Ensure you have Go installed on your system (Go 1.16 or later).
2.  **Clone the Repository**:
    ```bash
    git clone https://github.com/d3nd3/dota-report-timestamps
    cd dota-report-timestamps
    ```
3.  **Download Dependencies**:
    ```bash
    go mod download
    ```

## Web Interface

The tool provides a modern web interface for easy usage and batch processing.

### Building and Running

**Quick Start** (Recommended):
```bash
./run.sh
```
This script will build and run the server automatically.

**Manual Build and Run**:
1.  **Build the Server**:
    ```bash
    go build -o server ./cmd/server
    ```

2.  **Run the Server**:
    ```bash
    ./server
    ```

The server will start on `http://localhost:8081` by default.

**Open in Browser**:
Navigate to `http://localhost:8081` in your web browser.

### Configuration

Before using the tool, you need to configure:

1.  **Replay Directory**: Enter the full path to your Dota 2 replays directory.
    - Default location (Linux): `~/.steam/debian-installation/steamapps/common/dota 2 beta/game/dota/replays/`
    - Default location (Windows): `C:\Program Files (x86)\Steam\steamapps\common\dota 2 beta\game\dota\replays\`
    - Click "Save Settings" to save the configuration.

2.  **Stratz API Token** (Optional but highly recommended):
    - Get your API token from [Stratz API](https://stratz.com/api)
    - Enter the token in the "Stratz API Token" field
    - This enables fetching match history and finding replay URLs directly.

3.  **Steam API Key** (Optional but recommended for reliability):
    - Get your Steam Web API Key from [Steam Community](https://steamcommunity.com/dev/apikey)
    - Enter the key in the "Steam API Key" field
    - This serves as a fallback method to acquire replay salt/cluster info if Stratz fails.

4.  **Steam Bot Account** (Optional but most reliable):
    - The tool can log in to the Steam Game Coordinator (GC) using a Steam account to fetch replay salts directly from Valve.
    - This is the most reliable method as it bypasses API limits and works for very new matches where WebAPI might fail (500 error).
    - **Usage**: 
        - Use the "Steam GC Login" section in the web interface to log in.
        - Or set `STEAM_USER` and `STEAM_PASS` environment variables before running the server.
    - Note: You may need to provide a Steam Guard code via the UI if prompted.

### Features

#### Local Replays
- **View Replays**: Automatically lists all `.dem` files in your configured replay directory
- **Sort**: Toggle between newest and oldest replay files
- **Select**: Use "All" or "None" buttons to quickly select/deselect replays
- **Delete**: Remove selected replay files from your directory (permanent action)

#### Match History
- **Fetch Matches**: Enter a Steam ID and limit to fetch recent matches from Stratz
- **Auto Download**: Newly found matches are automatically queued for download (one at a time) as soon as the list loads
- **Progress**: A status pill above the list shows queued/downloading/completed states so you know when replays are ready
- **Add to List**: Click on any match ID to add it to your local replay selection for analysis (useful if a replay already exists)

#### Analysis
- **Select Replays**: Choose one or more replays from your local list
- **Target Options** (Optional):
    - **Steam ID**: Analyze reports for a specific player by their Steam ID
    - **Slot ID**: Analyze reports for a specific player by their slot (0-9)
- **Start Analysis**: Click "Start Analysis" to process selected replays
- **Results**: View detailed logs showing:
    - Total team reports vs enemy reports
    - Confirmed vs unconfirmed reports
    - Individual report timestamps and details

### Usage Example

1. Configure your replay directory and Stratz API token (if available)
2. Optionally fetch match history for a Steam ID to download recent replays
3. Select one or more replays from your local list
4. (Optional) Enter a target Steam ID or Slot ID to filter reports
5. Click "Start Analysis" and wait for results
6. Review the conclusion and detailed logs to see who reported whom and when

## Implementation Details

-   **Aspect Ratio**: The tool reads the aspect ratio from the replay to calculate button positions correctly.
-   **Scoreboard Variations**:
    -   When a player runs out of tips, the report button shifts 100 pixels to the left.
    -   Special hero assets (e.g., Wraith King Arcana) can alter the scoreboard layout.
-   **Mouse Tracking**: The tool tracks the mouse cursor position of players when they have the scoreboard open to detect clicks on the report button area.
-   **Replay Download**: The tool can automatically download replays from Valve's servers via Stratz API integration.
-   **Rate Limiting**: All OpenDota polling is throttled to ~30 requests per minute, and downloads are serialized to avoid hitting API limits.
-   **Fallback Mechanism**: If Stratz fails to provide replay info, the tool attempts to use Steam WebAPI (if configured) and finally falls back to OpenDota parsing.
-   **Replay Salt**: Valve's `IDOTA2Match_570/GetMatchDetails/v1` endpoint is broken for 7.36+ matches (500 or empty), so the only reliable salt sources are GC bots (e.g. go-dota2/node-dota2) or third parties like OpenDota/Stratz when they have processed the match.

## Troubleshooting

### Common Issues

-   **Missing Dependencies**: Run `go mod tidy` to ensure all modules are up to date.
-   **Replay Directory Not Found**: 
    -   Verify the path is correct and points to your Dota 2 replays folder
    -   Ensure the directory exists and contains `.dem` files
    -   On Linux, the default path is usually `~/.steam/debian-installation/steamapps/common/dota 2 beta/game/dota/replays/`
-   **Stratz/Steam API Errors**:
    -   Verify your API token/key is correct
    -   Check permissions and internet connection
-   **Replay File Not Found**:
    -   Make sure the replay file exists in your configured replay directory
    -   Replay files must be downloaded from Dota 2 or via the Match History feature
    -   Some older replays may no longer be available on Valve's servers
-   **Download Timeout**:
    -   Replay downloads can take several minutes, especially if the match needs to be parsed first via OpenDota
    -   The download will timeout after 10 minutes - check server logs for details
    -   Large replays or slow connections may require multiple attempts

### Server Logs

The server logs important information to help diagnose issues:
-   Configuration changes
-   Replay parsing progress
-   Download status
-   API errors

Check the terminal where you ran `./server` to see detailed logs.
