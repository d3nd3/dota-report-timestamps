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

2.  **Stratz API Token** (Optional but recommended):
    - Get your API token from [Stratz API](https://stratz.com/api)
    - Enter the token in the "Stratz API Token" field
    - The token is automatically saved and used for fetching match history and downloading replays
    - This enables the "Match History" feature to fetch recent matches

### Features

#### Local Replays
- **View Replays**: Automatically lists all `.dem` files in your configured replay directory
- **Sort**: Toggle between newest and oldest replay files
- **Select**: Use "All" or "None" buttons to quickly select/deselect replays
- **Delete**: Remove selected replay files from your directory (permanent action)

#### Match History
- **Fetch Matches**: Enter a Steam ID and limit to fetch recent matches from Stratz
- **Download Replays**: Click "Download" on any match to automatically download the replay file
- **Batch Download**: Use "Download All New" to download all matches that aren't already in your local directory
- **Add to List**: Click on any match ID to add it to your local replay list for analysis

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

## Troubleshooting

### Common Issues

-   **Missing Dependencies**: Run `go mod tidy` to ensure all modules are up to date.
-   **Replay Directory Not Found**: 
    -   Verify the path is correct and points to your Dota 2 replays folder
    -   Ensure the directory exists and contains `.dem` files
    -   On Linux, the default path is usually `~/.steam/debian-installation/steamapps/common/dota 2 beta/game/dota/replays/`
-   **Stratz API Errors**:
    -   Verify your API token is correct
    -   Check that your token has the necessary permissions
    -   Ensure you have an active internet connection
-   **Replay File Not Found**:
    -   Make sure the replay file exists in your configured replay directory
    -   Replay files must be downloaded from Dota 2 or via the Match History feature
    -   Some older replays may no longer be available on Valve's servers
-   **Download Timeout**:
    -   Replay downloads can take several minutes, especially if the match needs to be parsed first
    -   The download will timeout after 10 minutes - check server logs for details
    -   Large replays or slow connections may require multiple attempts

### Server Logs

The server logs important information to help diagnose issues:
-   Configuration changes
-   Replay parsing progress
-   Download status
-   API errors

Check the terminal where you ran `./server` to see detailed logs.
