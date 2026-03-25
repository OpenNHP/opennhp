import puppeteer from 'puppeteer'
import { execSync } from 'child_process'
import { mkdirSync, existsSync } from 'fs'
import path from 'path'

const URL = process.env.CAPTURE_URL || 'http://localhost:5180/arch-demo'
const OUT_DIR = path.resolve('recordings')
const DURATION = 8          // seconds to record
const FPS = 30
const WIDTH = 1200
const HEIGHT = 800

if (!existsSync(OUT_DIR)) mkdirSync(OUT_DIR)

;(async () => {
  console.log('Launching browser...')
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
  })

  const page = await browser.newPage()
  await page.setViewport({ width: WIDTH, height: HEIGHT, deviceScaleFactor: 2 })

  console.log('Loading page...')
  await page.goto(URL, { waitUntil: 'networkidle0', timeout: 30000 })

  // Wait for Vue SPA to hydrate and render the component
  console.log('Waiting for component to render...')
  // Debug: check what's on the page
  await new Promise(r => setTimeout(r, 5000))
  const html = await page.content()
  const hasWrapper = html.includes('arch-wrapper')
  const hasSvg = html.includes('<svg')
  const title = await page.title()
  console.log(`  Page title: "${title}", has arch-wrapper: ${hasWrapper}, has svg: ${hasSvg}`)
  if (!hasWrapper) {
    const bodyText = await page.evaluate(() => document.body.innerText.slice(0, 200))
    console.log(`  Body text: ${bodyText}`)
  }
  await page.waitForSelector('.arch-wrapper', { timeout: 15000 })
  // Extra wait for SVG animations to start
  await new Promise(r => setTimeout(r, 2000))

  // Take frames via screenshots
  const totalFrames = DURATION * FPS
  const frameDir = path.join(OUT_DIR, 'frames')
  if (!existsSync(frameDir)) mkdirSync(frameDir, { recursive: true })

  // Get the diagram element bounding box
  const box = await page.evaluate(() => {
    const el = document.querySelector('.arch-wrapper')
    if (!el) return null
    const rect = el.getBoundingClientRect()
    return { x: rect.x, y: rect.y, width: rect.width, height: rect.height }
  })

  if (!box) {
    console.error('Could not find .arch-wrapper element')
    await browser.close()
    process.exit(1)
  }

  // Add some padding around the diagram
  const pad = 20
  const clip = {
    x: Math.max(0, box.x - pad),
    y: Math.max(0, box.y - pad),
    width: box.width + pad * 2,
    height: box.height + pad * 2,
  }

  console.log(`Capturing ${totalFrames} frames at ${FPS}fps (${DURATION}s)...`)
  console.log(`Clip region: ${clip.width.toFixed(0)}x${clip.height.toFixed(0)}`)

  // Pause SVG animations and manually step through time.
  // This ensures the animation timeline is perfectly synced with the
  // output framerate, regardless of how long each screenshot takes.
  await page.evaluate(() => {
    const svg = document.querySelector('svg')
    if (svg) svg.pauseAnimations()
  })

  for (let i = 0; i < totalFrames; i++) {
    // Set the exact SVG animation time for this frame
    const t = i / FPS
    await page.evaluate((time) => {
      const svg = document.querySelector('svg')
      if (svg) svg.setCurrentTime(time)
    }, t)

    const frameNum = String(i).padStart(4, '0')
    await page.screenshot({
      path: path.join(frameDir, `frame-${frameNum}.png`),
      clip,
    })
    if (i % FPS === 0) process.stdout.write(`  ${Math.round(t)}s...`)
  }
  console.log(' done!')

  await browser.close()

  // Assemble video with ffmpeg
  const mp4Path = path.join(OUT_DIR, 'arch-diagram.mp4')
  const gifPath = path.join(OUT_DIR, 'arch-diagram.gif')
  const palettePath = path.join(OUT_DIR, 'palette.png')

  console.log('Encoding MP4...')
  execSync(
    `ffmpeg -y -framerate ${FPS} -i "${frameDir}/frame-%04d.png" ` +
    `-c:v libx264 -pix_fmt yuv420p -crf 18 -preset slow ` +
    `-vf "pad=ceil(iw/2)*2:ceil(ih/2)*2" ` +
    `"${mp4Path}"`,
    { stdio: 'inherit' }
  )

  console.log('Generating GIF (2-pass for quality)...')
  // Pass 1: generate optimized palette
  execSync(
    `ffmpeg -y -framerate ${FPS} -i "${frameDir}/frame-%04d.png" ` +
    `-vf "fps=${FPS},scale=800:-1:flags=lanczos,palettegen=max_colors=128:stats_mode=diff" ` +
    `"${palettePath}"`,
    { stdio: 'inherit' }
  )
  // Pass 2: encode GIF using palette
  execSync(
    `ffmpeg -y -framerate ${FPS} -i "${frameDir}/frame-%04d.png" -i "${palettePath}" ` +
    `-lavfi "fps=${FPS},scale=800:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=3" ` +
    `"${gifPath}"`,
    { stdio: 'inherit' }
  )

  // Clean up frames
  execSync(`rm -rf "${frameDir}" "${palettePath}"`)

  console.log(`\nDone!`)
  console.log(`  MP4: ${mp4Path}`)
  console.log(`  GIF: ${gifPath}`)

  // Print file sizes
  const mp4Size = execSync(`ls -lh "${mp4Path}" | awk '{print $5}'`).toString().trim()
  const gifSize = execSync(`ls -lh "${gifPath}" | awk '{print $5}'`).toString().trim()
  console.log(`  MP4 size: ${mp4Size}`)
  console.log(`  GIF size: ${gifSize}`)
})()
