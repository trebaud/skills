---
name: demo-gif
description: >
  Record polished demo GIFs of any terminal app (TUI or CLI) using VHS-scripted
  .tape files, verified frame-by-frame with screenshots. Use when the user asks
  to create a demo GIF/recording/screencast for a terminal program from a given
  set of instructions, "record a demo", make a terminal demo, or produce
  asciinema/VHS-style animations.
allowed-tools: Bash Read Write Edit
---

# Record demo GIFs for a terminal app

Produce a small set of deterministic, scripted GIFs of a terminal app from a
given set of instructions. Every GIF is a `.tape` file (VHS) replayed against a
fixed setup, so it is reproducible and survives UI changes — never hand-recorded.

**Core loop:** drive the app → screenshot key beats → Read the screenshots to
verify → adjust the tape → regenerate. Never ship a GIF you have not visually
inspected.

## 1. Install the toolchain (no sudo) — skip if already present

VHS needs `vhs` + `ttyd` + `ffmpeg`. If they are already on `PATH` (or in
`~/.local/vhs-tools/bin` from a prior run), reuse them and skip this step.
Otherwise run the bundled installer (idempotent — a no-op when tools exist):

```bash
bash scripts/setup-tools.sh   # installs into ~/.local/vhs-tools/bin
export PATH="$HOME/.local/vhs-tools/bin:$PATH"           # use the line it prints
```

Confirm `vhs --version`, `ttyd --version`, `ffmpeg -version` all work. (`go` and
`curl` are the only prerequisites. On macOS/non-x86_64, skip the script and use
`brew install vhs ttyd ffmpeg` instead — the installer fetches Linux x86_64 binaries.)

## 2. Know the app and prepare its input

- Make the binary runnable — build it and put it on `PATH`, or call it by
  absolute path inside the tape.
- Read its `--help` / key bindings so you know which keys drive which feature;
  the tape sends those keys literally.
- **Set the stage so the demo looks good.** Whatever the app needs to render
  something worth showing — sample files, a config, seed data, a fixture
  directory, command-line args — prepare it *reproducibly* (script the setup) and
  keep it **outside the app's own repo** (a sibling dir or `/tmp`). Avoid
  on-screen timestamps or random content so the same input yields the same GIF.

## 3. Calibrate sizing once

Render one throwaway tape that launches the app and takes a `Screenshot`. Read
the PNG and adjust until the layout is readable and fills the frame. Good
starting point: `FontSize 15`, `Width 1320`, `Height 820`, `Padding 14`.

## 4. Author the tapes

One tape per coherent story (aim for **2–4 GIFs**, ~12–25s each). Group related
features; do not cram everything into one.

```tape
Output "/abs/path/to/assets/demo.gif"   # paths MUST be quoted

Set Shell bash
Set FontSize 15
Set Width 1320
Set Height 820
Set Padding 14
Set TypingSpeed 70ms
# Optional custom theme so the padding matches the app's canvas color (avoids a
# border seam). Set "background" to the app's actual background.
Set Theme { "name": "app-dark", "background": "#0d1117", "foreground": "#c9d1d9", "cursor": "#c9d1d9" }

# Hide setup keystrokes (cd + PATH) from the recording.
Hide
Type 'export PATH="$HOME/.local/vhs-tools/bin:$PATH"; cd "$HOME/path/to/fixture"; clear'
Enter
Sleep 800ms
Show

Type "myapp"
Enter
Sleep 2s

# ── drive features; each key is sent as a real keypress ──
Type "s"          # single-key TUI binding
Sleep 1500ms
Type "jj"         # navigation
Ctrl+d            # special keys: Enter, Tab, Escape, Ctrl+x, Up/Down/Left/Right, Space
Sleep 1800ms

# Verification screenshot at each beat (PNG only — does NOT add gif frames):
Screenshot "/tmp/beat1.png"
```

Authoring rules / gotchas:

- **Quote every path** after `Output` and `Screenshot` — unquoted `/tmp/x.gif`
  fails the parser ("Invalid command: tmp").
- **`Type` with embedded double quotes** (e.g. `"$VAR"` in a shell command): wrap
  the VHS string in **single quotes** — `Type 'export PATH="..."'`.
- `Type "s"` sends the `s` key (capitals send shift); use named commands for
  non-printing keys.
- Pace with explicit `Sleep`s — 1.5–2.5s on a beat you want read; longer for a
  dense screen.
- Use `Hide`/`Show` to keep `cd`/`export`/`clear` out of frame.
- `vhs validate tape.tape` catches parse errors fast.

## 5. Generate, then verify by reading screenshots

```bash
vhs assets/tapes/demo.tape
```

For **every** `Screenshot`, `Read` the PNG and confirm the beat landed (right
view open, popup visible, content populated). Fix timing/keys and regenerate
until each beat is correct. This visual check is the heart of the skill — typed
keys often land a row off, or a `Sleep` catches a half-rendered frame.

## 6. Finalize — GIFs are the only deliverable

Once every beat is verified, do a final clean render and remove all scaffolding.
**The only artifacts left behind are the demo GIFs.**

- Move the verified GIFs to a conventional dir (`assets/` or `docs/`). Typical
  size ~0.5–1.5 MB each; mention total weight and offer to optimize if it matters.
- Delete everything used to generate them — the `.tape` files, screenshot PNGs,
  any throwaway calibration tape, and the temporary fixture/input dir:

  ```bash
  rm -f /tmp/beat*.png                  # verification screenshots
  rm -rf "$HOME/path/to/fixture"        # the reproducible input dir
  rm -f assets/tapes/*.tape && rmdir assets/tapes 2>/dev/null  # tape sources
  ```

- **Do NOT delete the installed toolchain** (`~/.local/vhs-tools/bin`). It is not
  a deliverable but it is reusable — leaving it means the next run skips the
  install in step 1. Only the per-demo scaffolding above is cleaned up.
- Confirm nothing but the GIFs remains (e.g. `git status` shows only the new
  `*.gif` files). If the user later wants to re-record, the tape is cheap to
  regenerate from this skill.
