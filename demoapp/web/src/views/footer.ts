// Shared page footer, mirroring the one on https://agent.opennhp.org/:
// a monospace public-IP line, an OpenNHP attribution, the sponsor credit,
// and a link to this demo app's source. The IP is fetched client-side
// from ipify (same pattern as the js-agent demo); the app's CSP
// connect-src must allow api.ipify.org / api6.ipify.org for it to load.

// Source-code call-out, rendered as its own section above the footer.
export function renderSourceSection(root: HTMLElement): void {
  const section = document.createElement('section');
  section.className = 'panel source-section';
  section.innerHTML = `
    <a href="https://github.com/OpenNHP/opennhp/tree/main/demoapp" target="_blank" rel="noopener noreferrer">
      <svg class="oauth-icon" viewBox="0 0 16 16" width="16" height="16" fill="currentColor" aria-hidden="true">
        <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.02.37-2.45-.49-2.6-.67-.09-.23-.48-.67-.82-.82-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.012 8.012 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
      </svg>
      View this demo app's source on GitHub
    </a>
  `;
  root.appendChild(section);
}

export function renderFooter(root: HTMLElement): void {
  const footer = document.createElement('footer');
  footer.className = 'footer';
  footer.innerHTML = `
    <div class="ip-display" id="ipDisplay">Detecting IP…</div>
    <div>Powered by <a href="https://github.com/OpenNHP/opennhp" target="_blank" rel="noopener"><strong>OpenNHP</strong></a> &mdash; Network-infrastructure Hiding Protocol</div>
    <div>Sponsored by: <a href="https://layerv.ai" target="_blank" rel="noopener">LayerV.ai</a></div>
  `;
  root.appendChild(footer);
  void fetchIP();
}

// Detect the visitor's public IPv4/IPv6 and show it in the footer.
async function fetchIP(): Promise<void> {
  const el = document.getElementById('ipDisplay');
  if (!el) return;
  let ipv4: string | null = null;
  let ipv6: string | null = null;
  try {
    const r = await fetch('https://api.ipify.org?format=json');
    if (r.ok) ipv4 = (await r.json()).ip;
  } catch (_) { /* ignore */ }
  try {
    const r = await fetch('https://api6.ipify.org?format=json');
    if (r.ok) ipv6 = (await r.json()).ip;
  } catch (_) { /* ignore */ }
  if (ipv4 || ipv6) {
    if (ipv4 && ipv6) el.textContent = `IPv4: ${ipv4}  |  IPv6: ${ipv6}`;
    else if (ipv4) el.textContent = `IPv4: ${ipv4}`;
    else el.textContent = `IPv6: ${ipv6}`;
  } else {
    el.textContent = 'IP detection unavailable';
  }
}
