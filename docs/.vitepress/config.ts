import { defineConfig } from 'vitepress'

// This VitePress setup is used exclusively for local rendering of the
// interactive Vue architecture diagram (ArchDiagram.vue), so that
// capture-arch.mjs can record it as a GIF via Puppeteer.
// It is NOT part of the CI/CD documentation pipeline — the live site
// continues to be built by Jekyll (.github/workflows/pages.yml).
// Exclude all markdown except the arch-demo page to avoid VitePress build
// errors from unconverted Jekyll pages. Add new VitePress pages here.
export default defineConfig({
  title: 'OpenNHP Documentation',
  description: 'Zero Trust Network-infrastructure Hiding Protocol',
  srcExclude: ['**/*.md', '!arch-demo.md', '!claw-dhp-demo.md'],

  themeConfig: {
    logo: '/images/logo12.png',
    nav: [
      { text: 'Architecture', link: '/arch-demo' },
      { text: 'OpenClaw + DHP', link: '/claw-dhp-demo' },
    ],
    sidebar: [
      {
        text: 'Visualization',
        items: [
          { text: 'Architecture Diagram', link: '/arch-demo' },
          { text: 'OpenClaw + DHP', link: '/claw-dhp-demo' },
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
