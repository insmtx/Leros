import { mkdir } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import sharp from 'sharp'

const currentDir = dirname(fileURLToPath(import.meta.url))
const appDir = resolve(currentDir, '..')
const sourceLogo = resolve(appDir, '../../packages/app-ui/assets/logo.svg')
const resourcesDir = join(appDir, 'resources')

await mkdir(resourcesDir, { recursive: true })
await sharp(await renderIcon(1024)).toFile(join(resourcesDir, 'icon.png'))

async function renderIcon(size) {
  // 中文注释：桌面图标直接使用透明底主体，并继续放大到接近满幅但避免裁边。
  const logoScale = size <= 64 ? 1 : 0.98
  const logoSize = Math.round(size * logoScale)
  const logoOffset = Math.round((size - logoSize) / 2)

  const logo = await sharp(sourceLogo)
    .resize(logoSize, logoSize, { fit: 'contain' })
    .png()
    .toBuffer()

  return sharp({
    create: {
      width: size,
      height: size,
      channels: 4,
      background: { r: 0, g: 0, b: 0, alpha: 0 },
    },
  })
    .composite([{ input: logo, left: logoOffset, top: logoOffset }])
    .png()
    .toBuffer()
}
