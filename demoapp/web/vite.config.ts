import { defineConfig } from 'vite';
import { resolve } from 'node:path';

export default defineConfig({
  // Map `@opennhp/agent` to the prebuilt local SDK. The Docker build
  // copies endpoints/js-agent/dist into node_modules/@opennhp/agent/dist
  // before running `vite build` (see Dockerfile.demoapp stage 1).
  // In local dev, run `cd ../../endpoints/js-agent && npm ci && npm run build`
  // first, then `npm run build` here. This avoids npm's file: dependency
  // resolution, which fails on `private: true` packages.
  resolve: {
    alias: {
      '@opennhp/agent': resolve(__dirname, 'node_modules/@opennhp/agent'),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    target: 'es2020',
  },
  server: {
    port: 5174,
    proxy: {
      // Proxy /api and /auth/oidc to the demo backend during dev so the SPA
      // doesn't need a same-origin policy to talk to the Go server.
      '/api': 'http://localhost:8081',
    },
  },
});
