import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';
import { readFileSync } from 'node:fs';
import { resolve } from 'path';

// Single source of truth shared with the Go binaries (see Makefile: VERSION =
// $(shell cat nhp/version/VERSION).$(TIMESTAMP)). Bump nhp/version/VERSION and
// both the binaries and this SDK pick up the new number on next build.
const sdkVersion = readFileSync(resolve(__dirname, '../../nhp/version/VERSION'), 'utf8').trim();

export default defineConfig({
  plugins: [
    dts({
      insertTypesEntry: true,
      rollupTypes: true,
    }),
  ],
  define: {
    __SDK_VERSION__: JSON.stringify(sdkVersion),
  },
  test: {
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      exclude: ['node_modules/', 'dist/', 'test/', '*.config.*'],
    },
  },
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'NHPAgent',
      formats: ['es', 'cjs'],
      fileName: (format) => `index.${format === 'es' ? 'js' : 'cjs'}`,
    },
    rollupOptions: {
      external: ['@noble/ciphers', '@noble/curves', '@noble/hashes'],
      output: {
        globals: {
          '@noble/ciphers': 'NobleCiphers',
          '@noble/curves': 'NobleCurves',
          '@noble/hashes': 'NobleHashes',
        },
      },
    },
    sourcemap: true,
    minify: false,
  },
});
