import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const frontendRoot = path.resolve(__dirname, '..')
const distDir = path.join(frontendRoot, 'dist')
const backendClientDir = path.resolve(frontendRoot, '..', 'backend', 'internal', 'webclient', 'dist')

if (!fs.existsSync(distDir)) {
  console.error('frontend dist directory not found, run vite build first')
  process.exit(1)
}

fs.rmSync(backendClientDir, { recursive: true, force: true })
fs.mkdirSync(backendClientDir, { recursive: true })
fs.cpSync(distDir, backendClientDir, { recursive: true })

console.log(`copied frontend build to ${backendClientDir}`)
