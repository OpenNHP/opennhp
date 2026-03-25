import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'OpenNHP Documentation',
  description: 'Zero Trust Network-infrastructure Hiding Protocol',
  srcExclude: ['**/zh-cn/**', 'about.md', 'agent_sdk.md', 'build.md', 'code.md', 'comparison.md', 'cryptography.md', 'deploy.md', 'dhp_quick_start.md', 'features.md', 'index.md', 'nhp_quick_start.md', 'server_plugin.md', 'README.md'],  // TODO: remove srcExclude after migrating Jekyll markdown syntax

  themeConfig: {
    logo: '/images/logo12.png',
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Features', link: '/features' },
      { text: 'Quick Start', link: '/nhp_quick_start' },
    ],
    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Overview', link: '/' },
          { text: 'About', link: '/about' },
          { text: 'Features', link: '/features' },
        ]
      },
      {
        text: 'Guide',
        items: [
          { text: 'Build', link: '/build' },
          { text: 'Deploy', link: '/deploy' },
          { text: 'NHP Quick Start', link: '/nhp_quick_start' },
          { text: 'DHP Quick Start', link: '/dhp_quick_start' },
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'Architecture', link: '/code' },
          { text: 'Cryptography', link: '/cryptography' },
          { text: 'Agent SDK', link: '/agent_sdk' },
          { text: 'Server Plugin', link: '/server_plugin' },
          { text: 'Comparison', link: '/comparison' },
        ]
      }
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
