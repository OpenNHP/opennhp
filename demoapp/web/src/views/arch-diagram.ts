// Architecture diagram for the login page. Renders a compact, CSP-safe
// (no external assets) flow showing how a knock travels from the browser
// through nhp-relay and nhp-server to nhp-ac, and back as access info.
//
// The NHP protocol has two client-driven phases the demo exercises:
//   - Register: the agent's public key is sent to nhp-server (NHP_REG) so
//     the server can authenticate later knocks.
//   - Knock: an encrypted NHP_KNK proves the agent owns the registered key;
//     the server authorizes and asks nhp-ac to open the hidden port.
//
// Numbered steps below the node strip trace one knock end-to-end.

export function renderArchDiagram(root: HTMLElement): void {
  root.innerHTML = `
    <details class="arch" open>
      <summary>How the OpenNHP demo works</summary>
      <div class="arch-body">
        <div class="arch-nodes">
          <div class="arch-node">
            <div class="arch-node-title">Browser</div>
            <div class="arch-node-sub">SPA + js-agent</div>
          </div>
          <div class="arch-arrow">&rarr;</div>
          <div class="arch-node">
            <div class="arch-node-title">nhp-relay</div>
            <div class="arch-node-sub">routes packets</div>
          </div>
          <div class="arch-arrow">&rarr;</div>
          <div class="arch-node arch-node-key">
            <div class="arch-node-title">nhp-server</div>
            <div class="arch-node-sub">auth &amp; authz</div>
          </div>
          <div class="arch-arrow">&rarr;</div>
          <div class="arch-node">
            <div class="arch-node-title">nhp-ac</div>
            <div class="arch-node-sub">firewall</div>
          </div>
          <div class="arch-arrow">&rarr;</div>
          <div class="arch-node">
            <div class="arch-node-title">Resource</div>
            <div class="arch-node-sub">hidden service</div>
          </div>
        </div>

        <ol class="arch-steps">
          <li>
            <strong>NHP-REG</strong> (once, at sign-up): the browser's js-agent
            sends the user's public key to nhp-server via the relay, so later
            knocks can be authenticated.
          </li>
          <li>
            <strong>NHP_KNK</strong>: the agent encrypts a knock and sends it to
            nhp-server through nhp-relay. The relay only routes opaque packets,
            so it learns nothing about the request.
          </li>
          <li>
            <strong>NHP_AOP</strong>: nhp-server validates the knock and asks
            nhp-ac to open the hidden port for this agent.
          </li>
          <li>
            <strong>NHP_ART</strong>: nhp-ac opens the firewall rule and reports
            the result back to nhp-server.
          </li>
          <li>
            <strong>NHP_ACK</strong>: nhp-server returns the access info (host,
            port, TTL) to the agent through the relay.
          </li>
          <li>
            The browser now reaches the protected resource directly through
            nhp-ac's opened port &mdash; until the rule expires, the service
            stays invisible to everyone else.
          </li>
        </ol>
      </div>
    </details>
  `;
}
