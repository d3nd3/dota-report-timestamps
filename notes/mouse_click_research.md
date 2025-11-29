# Mouse Click Detection Research

## Current Implementation

The parser currently **infers** mouse clicks by:
1. Tracking cursor position (`m_iCursor.0000` and `m_iCursor.0001`) from `CDOTAPlayerController` entity
2. Detecting when cursor hovers over report buttons (tracking hover duration)
3. Detecting when cursor enters the confirm box area (X: 968-1159, Y: 844-889)
4. Inferring a click occurred when cursor is in confirm box within 120 ticks of hovering over a report button

This is an **indirect method** - it doesn't detect actual mouse button presses, but rather infers clicks based on cursor position and timing.

## Direct Mouse Click Detection - Possible but Complex

### Available Data Structures

1. **CDemoUserCmd** - Available via `p.Callbacks.OnCDemoUserCmd()`
   - Contains `cmd_number` (int32) and `data` (raw bytes)
   - The `data` field contains serialized user command data

2. **CBaseUserCmdPB** - Structure that can be parsed from CDemoUserCmd.data
   - Contains `buttons_pb` field of type `CInButtonStatePB`
   - `CInButtonStatePB` has:
     - `buttonstate1` (uint64)
     - `buttonstate2` (uint64) 
     - `buttonstate3` (uint64)
   - These button states contain information about which mouse/input buttons are pressed

### Implementation Challenges

1. **Data Parsing**: The `CDemoUserCmd.data` field is raw bytes that need to be deserialized into `CBaseUserCmdPB` structure
2. **Button State Interpretation**: The button state bitfields need to be decoded to identify which buttons are pressed (left click, right click, etc.)
3. **Player Association**: User commands need to be associated with specific players (may require additional tracking)
4. **UI vs Game Clicks**: Need to distinguish between in-game clicks (unit selection, movement) vs UI clicks (report button clicks)

### Current Approach Advantages

The current cursor-based approach has advantages:
- **Simpler**: No need to parse complex user command structures
- **UI-Specific**: Directly tracks UI interactions (cursor position on scoreboard)
- **Reliable**: Works well for detecting report button interactions
- **Already Working**: The current implementation successfully detects report clicks

### Recommendations

1. **For Report Detection**: The current cursor-based approach is sufficient and recommended
   - It's simpler and more reliable for UI interactions
   - Mouse clicks in user commands are primarily for in-game actions, not UI

2. **For General Mouse Click Detection**: If you need to detect all mouse clicks:
   - Implement `OnCDemoUserCmd` callback
   - Parse the `data` field into `CBaseUserCmdPB` 
   - Decode `buttons_pb.buttonstate1/2/3` bitfields
   - Map button states to specific mouse buttons (left/right/middle)
   - Associate commands with players

3. **Hybrid Approach**: Could combine both:
   - Use cursor position for UI interactions (reports)
   - Use user commands for in-game mouse clicks (unit selection, attacks)

## Code References

- Current cursor tracking: `pkg/parser/parser.go:548-551`
- Confirm box detection: `pkg/parser/parser.go:591-607`
- Available callback: `vendor/github.com/dotabuff/manta/callbacks.go:352-355`
- User command structure: `vendor/github.com/dotabuff/manta/dota/usercmd.proto:24-42`
- Button state structure: `vendor/github.com/dotabuff/manta/dota/usercmd.proto:8-12`

## Example Implementation (Conceptual)

If you wanted to attempt direct mouse click detection, you would need to:

```go
p.Callbacks.OnCDemoUserCmd(func(cmd *dota.CDemoUserCmd) error {
    // cmd.GetData() returns []byte - this needs to be parsed
    // The data format is not standard protobuf - it's a custom binary format
    // You would need to:
    // 1. Parse the raw bytes into CBaseUserCmdPB structure
    // 2. Extract buttons_pb field
    // 3. Decode buttonstate1/2/3 bitfields to identify which buttons are pressed
    
    // Note: The parsing of cmd.GetData() is non-trivial and may require
    // reverse engineering the binary format or finding documentation
    return nil
})
```

**Important Note**: The `CDemoUserCmd.data` field is raw binary data that may not be in standard protobuf format. Parsing it correctly would require understanding the Source engine's user command serialization format, which is not well-documented.

## Conclusion

**Direct mouse click detection is technically possible** via `CDemoUserCmd`, but:
- Requires parsing raw binary data
- More complex than current approach
- May not be necessary for report detection (current method works well)
- Would be useful if you need to detect in-game mouse clicks (unit selection, attacks, etc.)

For the current use case (detecting report button clicks), the cursor-based inference method is the better choice.

