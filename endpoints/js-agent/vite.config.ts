import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';
import { resolve } from 'path';

export default defineConfig({
  plugins: [
    dts({
      insertTypesEntry: true,
      rollupTypes: true,
    }),
  ],
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
