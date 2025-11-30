# Features Guide

This document explains the advanced features available in the Dota 2 Report Timestamp Tool.

## 1. Fatal Game Search

### What is it?

The **Fatal Game Search** feature helps you find "fatal" games - ranked matches that occur immediately after a single draft (low priority) match. This is useful for analyzing behavior patterns when players transition from low priority back to ranked matches.

### How it works

The feature searches through a player's match history to find sequences where:
1. A **single draft** (low priority) match occurred
2. Followed by a **ranked** match (the "fatal" match)

It can find multiple such sequences up to a specified depth.

### How to use it

1. **Prerequisites**: You must be logged into Steam through the GUI (Steam GC Login section)

2. **Access the feature**: 
   - In the sidebar, find the "Fatal Game Search" section
   - The Steam ID field will auto-fill if you have a profile selected

3. **Configure parameters**:
   - **Steam ID**: The Steam ID of the player to search (auto-filled from selected profile)
   - **Max Depth**: Number of low priority sessions to search (1-10)
     - Depth 1 = Find the most recent low priority → ranked sequence
     - Depth 2 = Find the 2 most recent sequences
     - And so on...
   - **Games Per Fatal**: Number of ranked games to download before each singledraft match (1-15)
     - This downloads additional ranked matches that occurred before the singledraft
     - Useful for getting more context around the low priority event

4. **Run the search**:
   - Click "Find Fatal Games"
   - Wait for the search to complete (may take a minute)
   - Results will show each fatal match with its associated singledraft match

5. **Download matches**:
   - Click "Download to Fatal" on individual matches
   - Or click "Download All to Fatal" to download all found matches
   - Downloads are saved to a special "fatal" directory organized by date

### Example Use Case

If a player got low priority and you want to analyze their behavior:
- Set Max Depth to 1 to find the most recent occurrence
- Set Games Per Fatal to 5 to download 5 ranked games before the singledraft
- This gives you context: what happened in ranked that led to low priority, and how they played after returning

---

## 2. Report Cluster Verifier

### What is it?

The **Report Cluster Verifier** downloads a cluster of ranked matches (up to 15) starting from a specific match ID. This is used to verify and analyze report clusters - groups of matches that contribute to a player's conduct scorecard.

### How it works

When you validate a report card, the tool:
1. Finds the specified match in the player's history
2. Collects up to 15 ranked matches starting from that match (going backwards in time)
3. Downloads all matches that don't already exist locally
4. Saves them to a dedicated `reportcards` directory

### How to use it

1. **Prerequisites**: 
   - You must be logged into Steam through the GUI
   - You need a match ID and the player's Steam Account ID

2. **Access the feature**:
   - The feature appears in the **Conduct Scorecard** banner at the top of the page
   - This banner appears when viewing match results for a specific player

3. **Two validation modes**:

   **a) Validate Report Card** (Standard):
   - Downloads up to 15 ranked matches starting from the specified match ID
   - Goes backwards in time from the match
   - Includes the starting match itself if it's ranked
   - Skips single draft and turbo matches
   - Use this to verify a complete report cluster

   **b) Validate Report Card Current** (Current):
   - Downloads ranked matches that occurred AFTER the specified match ID
   - Goes forward in time from the match
   - Only downloads if there are fewer than 15 matches already
   - Use this to get matches that happened after a specific point

4. **Run validation**:
   - Click either "Validate Report Card" or "Current" button
   - The process may take several minutes as it:
     - Fetches match history from Steam
     - Downloads missing replays
     - Shows progress in the status message
   - Results show: downloaded count, skipped count, errors, and save location

5. **View results**:
   - All downloaded matches are saved to: `replays/reportcards/{matchId}/`
   - You can then analyze these matches using the main parser

### Example Use Case

If you see a conduct scorecard showing reports in match 123456789:
- Click "Validate Report Card" 
- The tool downloads match 123456789 and up to 14 ranked matches before it
- You can then parse these matches to verify the report data matches the scorecard
- This helps validate whether the conduct system is working correctly

### Important Notes

- Both features require an active Steam connection through the GUI
- Downloads may take time depending on match availability
- Some matches may be queued for parsing by Valve (you'll see a "queued" status)
- The tool automatically skips matches that already exist locally

