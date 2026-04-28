---
layout: page
title: Capture Diagram Animations
nav_order: 10
permalink: /capture-diagrams/
---

# Capture Diagram Animations
{: .fs-9 }

Record the interactive SVG architecture diagrams as GIF and MP4 files.
{: .fs-6 .fw-300 }

---

## Overview

The project provides two capture scripts that use Puppeteer to record SVG animations from the VitePress documentation site and encode them into GIF and MP4 via ffmpeg.

| Script | Target Page | Output Files |
|---|---|---|
| `capture-arch.mjs` | `/arch-demo` | `arch-diagram.mp4`, `arch-diagram.gif` |
| `capture-claw-dhp.mjs` | `/claw-dhp-demo` | `claw-dhp-diagram.mp4`, `claw-dhp-diagram.gif` |

## Prerequisites

- **Node.js** (v18+)
- **Chrome or Chromium** — auto-detected at common paths, or set the `CHROME_PATH` environment variable
- **ffmpeg** — must be available on `PATH`

## Install Dependencies

```bash
cd docs
npm install
```

## Usage

### Step 1: Start the VitePress Dev Server

```bash
cd docs
npm run dev -- --port 5180
```

Keep this terminal running.

### Step 2: Run the Capture Script

Open a new terminal:

```bash
cd docs

# Capture the NHP architecture diagram
node capture-arch.mjs

# Capture the OpenClaw + DHP architecture diagram
node capture-claw-dhp.mjs
```

Output files are written to `docs/recordings/`.

## Configuration

Both scripts support environment variables for customization:

| Variable | Description | Default |
|---|---|---|
| `CAPTURE_URL` | Full URL of the target page | `http://localhost:5180/<page>` |
| `CHROME_PATH` | Path to Chrome/Chromium executable | Auto-detected |

Example:

```bash
CHROME_PATH=/usr/bin/chromium node capture-arch.mjs
```

## Script Parameters

Default values are defined at the top of each script and can be edited directly:

| Parameter | `capture-arch.mjs` | `capture-claw-dhp.mjs` |
|---|---|---|
| Viewport | 1200 x 800 | 1920 x 1080 |
| Duration | 8 seconds | 8 seconds |
| FPS | 30 | 30 |
| GIF scale width | 800px | 1920px |
| GIF palette colors | 128 | 256 |

## How It Works

1. Puppeteer launches headless Chrome and opens the target VitePress page
2. Waits for the Vue component to render (`.arch-wrapper` or `.claw-dhp-wrapper`)
3. Pauses SVG animations and steps through the timeline frame-by-frame via `svg.setCurrentTime()`
4. Takes a screenshot for each frame (240 frames = 8s at 30fps)
5. Encodes the PNG sequence into MP4 (H.264) and GIF (two-pass with optimized palette) using ffmpeg
6. Cleans up temporary frame files
