// Architecture overview for the login page.
//
// Replaces an earlier inline text breakdown with the screenshot shipped at
// docs/images/demoapp/demoapp.png. Vite serves anything under web/public/
// from the site root, which lets the embedded binary (built with the
// `webdist` tag) include it via the same embed.FS that holds the rest of
// the SPA — no extra fetch, no CSP hole for cross-origin assets.

export function renderArchDiagram(root: HTMLElement): void {
  root.innerHTML = `
    <details class="arch" open>
      <summary>How the OpenNHP demo works</summary>
      <div class="arch-body">
        <img
          class="arch-diagram-image"
          src="/demoapp.png"
          alt="OpenNHP demo architecture: browser, nhp-relay, nhp-server, nhp-ac, and the protected resource, with the NHP-REG / NHP_KNK / NHP_AOP / NHP_ART / NHP_ACK message flow"
        />
      </div>
    </details>
  `;
}