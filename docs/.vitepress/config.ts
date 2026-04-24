import { defineConfig } from 'vitepress'

// This VitePress setup exists solely so tools/capture-*.mjs can render
// the interactive Vue architecture diagrams (ArchDiagram.vue and
// ClawDhpDiagram.vue) and screenshot them into GIFs. It is NOT part of
// the CI/CD documentation pipeline — the live site continues to be
// built by Jekyll (.github/workflows/pages.yml).
//
// The two pages VitePress serves live under .vitepress/pages/ (outside
// Jekyll's source tree) so Jekyll never sees them and VitePress has a
// clean, minimal source root to index.
export default defineConfig({
  title: 'OpenNHP Documentation',
  description: 'Zero Trust Network-infrastructure Hiding Protocol',
  srcDir: '.vitepress/pages',

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
