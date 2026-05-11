/**
 * HTTP Relay Transport for Browser Environments
 *
 * Forwards NHP packets to the NHP Server via an intermediary HTTP relay
 * service instead of a direct UDP / WebSocket connection.
 *
 * Flow:
 *   NHPAgent  ──send(KNK)──▶  HttpRelayTransport
 *                              │  POST /relay (binary body = raw KNK packet)
 *                              ▼
 *                         nhp-relay service
 *                              │  UDP NHLY frame (realClientIP + KNK)
 *                              ▼
 *                         NHP Server
 *                              │  UDP ACK/COK (encrypted to original agent)
 *                              ▼
 *                         nhp-relay service
 *                              │  HTTP 200 (binary body = raw ACK/COK bytes)
 *                              ▼
 *   NHPAgent  ◀──message────  HttpRelayTransport
 *
 * The transport is "virtually connected" as soon as it's created — there is no
 * persistent connection to maintain (each send() call opens and closes one
 * HTTP request).  The connect/disconnect methods are no-ops provided for
 * interface compatibility.
 */

import type { Logger, TransportEvent, EventHandler } from '../types.js';

const noopLogger: Logger = {
  error: () => {},
  info: () => {},
  debug: () => {},
};

/** Configuration for the HTTP relay transport */
export interface HttpRelayTransportConfig {
  /**
   * Full URL of the relay endpoint, e.g. "https://relay.example.com/relay"
   * or "http://localhost:8080/relay".
   *
   * The cluster ID (if any) is appended automatically; supplying a URL that
   * already includes a cluster ID is supported but discouraged because it
   * bypasses the per-call clusterId option below.
   */
  relayUrl: string;

  /**
   * Cluster ID (11-char base64url fingerprint of the target nhp-server's
   * public key). When set, requests go to `${relayUrl}/${clusterId}`.
   *
   * When omitted, the request hits `relayUrl` directly and the relay uses
   * its configured `defaultClusterId`. Browsers targeting a relay that
   * fronts multiple clusters MUST set this; single-cluster relays accept
   * either form.
   */
  clusterId?: string;

  /**
   * Request timeout in milliseconds (default: 10000).
   * Should be ≥ the relay's udpTimeoutMs setting.
   */
  timeoutMs?: number;

  /** Optional logger; if omitted, the transport stays silent in production. */
  logger?: Logger;
}

/**
 * HTTP Relay transport — sends NHP packets over HTTPS and receives
 * ACK/COK responses in the HTTP response body.
 */
export class HttpRelayTransport {
  // relayUrl and timeoutMs are always populated (default applied below);
  // clusterId and logger remain optional and inherit their original types.
  private readonly config: {
    relayUrl: string;
    timeoutMs: number;
    clusterId?: string;
    logger?: Logger;
  };
  private readonly eventHandlers: Map<TransportEvent, Set<EventHandler>> = new Map();
  private readonly log: Logger;
  private connected = false;

  constructor(config: HttpRelayTransportConfig) {
    this.config = {
      timeoutMs: 10_000,
      ...config,
    };
    this.log = config.logger ?? noopLogger;
  }

  // ─── Transport interface ──────────────────────────────────────────────────

  /** Always resolves immediately — no persistent connection needed. */
  async connect(): Promise<void> {
    this.connected = true;
    // Emit 'open' asynchronously so callers can attach handlers after connect().
    setTimeout(() => this.emit('open', undefined), 0);
  }

  /** Marks the transport as disconnected. */
  disconnect(): void {
    this.connected = false;
    this.emit('close', undefined);
  }

  /** Returns true if connect() has been called and disconnect() has not. */
  isConnected(): boolean {
    return this.connected;
  }

  /**
   * Sends an NHP packet to the relay and delivers the response via the
   * 'message' event.
   *
   * @throws Error if the relay returns a non-200 status or the request times
   *         out.
   */
  async send(data: Uint8Array): Promise<void> {
    if (!this.connected) {
      throw new Error('[HttpRelayTransport] not connected');
    }

    let respBytes: Uint8Array;
    try {
      respBytes = await this.postToRelay(data);
    } catch (err) {
      this.emit('error', err);
      throw err;
    }

    this.emit('message', { data: respBytes });
  }

  on(event: TransportEvent, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  off(event: TransportEvent, handler: EventHandler): void {
    this.eventHandlers.get(event)?.delete(handler);
  }

  // ─── Internal ─────────────────────────────────────────────────────────────

  private async postToRelay(packet: Uint8Array): Promise<Uint8Array> {
    const { relayUrl, clusterId, timeoutMs } = this.config;
    // Suffix the cluster ID as a path segment when provided. We avoid an
    // explicit `URL` constructor here so callers can pass relative URLs in
    // odd test setups.
    const targetUrl = clusterId
      ? `${relayUrl.replace(/\/+$/, '')}/${encodeURIComponent(clusterId)}`
      : relayUrl;

    // Always copy to a clean ArrayBuffer — packet may be a view into a larger buffer
    const body = new Uint8Array(packet).buffer;

    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), timeoutMs);

    let response: Response;
    try {
      response = await fetch(targetUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/octet-stream' },
        body: body as ArrayBuffer,
        signal: controller.signal,
      });
    } catch (err) {
      clearTimeout(timer);
      if ((err as Error).name === 'AbortError') {
        throw new Error(`[HttpRelayTransport] request to ${targetUrl} timed out after ${timeoutMs}ms`);
      }
      throw new Error(`[HttpRelayTransport] fetch failed: ${(err as Error).message}`);
    } finally {
      clearTimeout(timer);
    }

    if (!response.ok) {
      const errBody = await response.text().catch(() => '');
      throw new Error(
        `[HttpRelayTransport] relay returned HTTP ${response.status}: ${errBody.trim() || response.statusText}`
      );
    }

    const buf = await response.arrayBuffer();
    const respBytes = new Uint8Array(buf);
    this.log.debug(`[HttpRelayTransport] sent ${packet.byteLength}B, received ${respBytes.byteLength}B (HTTP ${response.status})`);
    return respBytes;
  }

  private emit(event: TransportEvent, data: unknown): void {
    this.eventHandlers.get(event)?.forEach((h) => h(data));
  }
}
