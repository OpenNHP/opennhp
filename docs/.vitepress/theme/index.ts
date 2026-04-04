import DefaultTheme from 'vitepress/theme'
import ArchDiagram from '../components/ArchDiagram.vue'
import ClawDhpDiagram from '../components/ClawDhpDiagram.vue'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('ArchDiagram', ArchDiagram)
    app.component('ClawDhpDiagram', ClawDhpDiagram)
  }
}
