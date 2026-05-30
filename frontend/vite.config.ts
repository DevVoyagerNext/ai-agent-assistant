import { createRequire } from 'node:module'
import { createReadStream, existsSync, readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, join, relative, resolve, sep } from 'node:path'
import { defineConfig, type Connect, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'

const require = createRequire(import.meta.url)
const vditorRoot = dirname(require.resolve('vditor/package.json'))
const vditorDistRoot = join(vditorRoot, 'dist')
const vditorBuildEntries = ['css', 'images', 'js', 'index.css', 'method.min.js']

const mimeTypes: Record<string, string> = {
  '.css': 'text/css; charset=utf-8',
  '.gif': 'image/gif',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.map': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.svg': 'image/svg+xml; charset=utf-8',
  '.ttf': 'font/ttf',
  '.wasm': 'application/wasm',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2'
}

const getMimeType = (filePath: string) => mimeTypes[extname(filePath).toLowerCase()] || 'application/octet-stream'

const resolveVditorAsset = (url?: string) => {
  if (!url?.startsWith('/vditor/')) return null

  const requestPath = decodeURIComponent(url.split('?')[0].replace(/^\/vditor\//, ''))
  const filePath = resolve(vditorRoot, requestPath)
  const rootPrefix = vditorRoot.endsWith(sep) ? vditorRoot : `${vditorRoot}${sep}`

  if (filePath !== vditorRoot && !filePath.startsWith(rootPrefix)) {
    return null
  }

  return filePath
}

const vditorAssetMiddleware: Connect.NextHandleFunction = (req, res, next) => {
  const filePath = resolveVditorAsset(req.url)
  if (!filePath || !existsSync(filePath) || !statSync(filePath).isFile()) {
    next()
    return
  }

  res.statusCode = 200
  res.setHeader('Content-Type', getMimeType(filePath))
  createReadStream(filePath).pipe(res)
}

const vditorStaticPlugin = (): Plugin => ({
  name: 'local-vditor-static-assets',
  configureServer(server) {
    server.middlewares.use(vditorAssetMiddleware)
  },
  configurePreviewServer(server) {
    server.middlewares.use(vditorAssetMiddleware)
  },
  generateBundle() {
    const emitAsset = (filePath: string) => {
      const relativePath = relative(vditorDistRoot, filePath).split(sep).join('/')
      this.emitFile({
        type: 'asset',
        fileName: `vditor/dist/${relativePath}`,
        source: readFileSync(filePath)
      })
    }

    const walk = (dir: string) => {
      for (const entry of readdirSync(dir)) {
        const filePath = join(dir, entry)
        const stats = statSync(filePath)

        if (stats.isDirectory()) {
          walk(filePath)
          continue
        }

        emitAsset(filePath)
      }
    }

    for (const entry of vditorBuildEntries) {
      const filePath = join(vditorDistRoot, entry)
      if (!existsSync(filePath)) {
        continue
      }

      if (statSync(filePath).isDirectory()) {
        walk(filePath)
      } else {
        emitAsset(filePath)
      }
    }
  }
})

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vditorStaticPlugin()],
  server: {
    port: 5173,
    strictPort: true,
  },
})
