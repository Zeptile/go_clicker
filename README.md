# go_clicker

A cross-platform auto clicker with color-filtered clicking. Only clicks when the pixel under your cursor matches target colors — useful for targeting specific game entities.

## Requirements

**All platforms:** Go 1.24+

**Linux (Wayland/Hyprland):**
- `ydotool` + `ydotoold` daemon for clicking
- `grim` for screen capture
- `slurp` for region selection (colorpicker)
- `hyprctl` for cursor position

**Linux (X11) / macOS / Windows:** No extra dependencies — uses [robotgo](https://github.com/go-vgo/robotgo).

## Build

```bash
go build -o go_clicker .
go build -o colorpicker ./cmd/colorpicker
```

## Usage

### Basic auto clicker

```bash
# Fixed interval (default 50ms)
./go_clicker

# Random interval between 0-500ms
./go_clicker --random --randomIntervalEnd 500
```

### Color-filtered clicking

**Step 1:** Capture target colors with the colorpicker tool.

```bash
# Single capture
./colorpicker colors.json

# Multiple captures (appends to file, deduplicates)
./colorpicker 3 colors.json
```

On Wayland, this opens `slurp` for drag-to-select. On X11/macOS/Windows, it uses a countdown timer.

**Step 2:** Run the clicker with the colors file.

```bash
./go_clicker --colorsFile colors.json
./go_clicker --colorsFile colors.json --colorTolerance 50 --debug
```

The clicker only clicks when the pixel under your cursor matches any color in the file (within tolerance).

### Controls

**Linux / macOS:** Uses Unix signals.

```bash
kill -USR1 $(pgrep go_clicker)   # toggle on/off
kill -USR2 $(pgrep go_clicker)   # sample color under cursor
```

**Hyprland:** Add to `~/.config/hypr/bindings.conf`:

```
bind = CTRL SHIFT, P, exec, pkill -USR1 go_clicker
bind = CTRL SHIFT, O, exec, pkill -USR2 go_clicker
```

**Windows:** Global hotkeys Ctrl+Shift+P (toggle) and Ctrl+Shift+O (sample).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--random` | `false` | Randomize click interval |
| `--intervalMs` | `50` | Click interval in normal mode (ms) |
| `--randomIntervalEnd` | `100` | Max random interval (ms) |
| `--color` | `false` | Enable color-filtered clicking |
| `--colorsFile` | | Path to colors.json (implies `--color`) |
| `--colorTolerance` | `30` | RGB tolerance per channel (0-255) |
| `--debug` | `false` | Print debug info (pixel colors, mismatches) |

## Platform Support

| | Linux (Wayland) | Linux (X11) | macOS | Windows |
|---|---|---|---|---|
| Clicking | ydotool | robotgo | robotgo | robotgo |
| Cursor position | hyprctl | robotgo | robotgo | robotgo |
| Pixel color | grim | robotgo | robotgo | robotgo |
| Region capture | slurp + grim | robotgo | robotgo | robotgo |

Linux auto-detects Wayland vs X11 at runtime via `$WAYLAND_DISPLAY` and `$XDG_SESSION_TYPE`.
