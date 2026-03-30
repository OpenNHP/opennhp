import { defineConfig } from 'vitepress'

// This VitePress setup is used exclusively for local rendering of the
// interactive Vue architecture diagram (ArchDiagram.vue), so that
// capture-arch.mjs can record it as a GIF via Puppeteer.
// It is NOT part of the CI/CD documentation pipeline — the live site
// continues to be built by Jekyll (.github/workflows/pages.yml).
// Existing Jekyll pages are excluded to avoid VitePress build errors.
export default defineConfig({
  title: 'OpenNHP Documentation',
  description: 'Zero Trust Network-infrastructure Hiding Protocol',
  srcExclude: ['**/zh-cn/**', 'about.md', 'agent_sdk.md', 'build.md', 'code.md', 'comparison.md', 'cryptography.md', 'deploy.md', 'dhp_quick_start.md', 'features.md', 'index.md', 'nhp_quick_start.md', 'server_plugin.md', 'README.md'],

  themeConfig: {
    logo: '/images/logo12.png',
    nav: [
      { text: 'Architecture', link: '/arch-demo' },
    ],
    sidebar: [
      {
        text: 'Visualization',
        items: [
          { text: 'Architecture Diagram', link: '/arch-demo' },
        ]
      },
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/OpenNHP/opennhp' }
    ],
    footer: {
      message: 'Released under the Apache 2.0 License.',
      copyright: 'Copyright © 2024 OpenNHP Open Source Project.'
    },
    editLink: {
      pattern: 'https://github.com/OpenNHP/opennhp/edit/main/docs/:path'
    },
    search: {
      provider: 'local'
    }
  }
})
