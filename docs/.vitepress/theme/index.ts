import DefaultTheme from 'vitepress/theme'
import ArchDiagram from '../components/ArchDiagram.vue'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('ArchDiagram', ArchDiagram)
  }
}
