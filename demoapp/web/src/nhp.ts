// nhp.ts — wraps js-agent for the Demo's two flows: registration and
// resource knock. The private key is delivered by the backend over HTTPS,
// held in memory only, and dropped once each flow completes.

import { NHPAgent } from '@opennhp/agent';
import type { NhpEndpointConfig } from './api.js';

export interface RegistrationResult {
  rakOk: boolean;
  expiresAt?: number;
}

export interface KnockOutcome {
  success: boolean;
  resourceHost?: string;
  error?: string;
}

export interface AgentHandle {
  agent: NHPAgent;
  /** Close the agent and wipe the in-memory private key reference. */
  dispose: () => void;
}

/**
 * Build an NHPAgent instance from backend-issued credentials. The private
 * key stays in `agent.keyPair.privateKey` for the lifetime of the agent;
 * callers must call `dispose()` to close transports and clear the chain
 * keys (the underlying string is no longer referenced but is not zeroed).
 */
export async function createAgent(
  privateKey: string,
  cfg: NhpEndpointConfig,
  deviceId: string,
  logLevel: 'silent' | 'error' | 'info' | 'debug' = 'error',
): Promise<AgentHandle> {
  const agent = new NHPAgent({
    cipherScheme: cfg.cipherScheme,
    transport: 'relay',
    relayUrl: cfg.relayUrl,
    privateKey,
    logLevel,
  });
  await agent.init();
  agent.setIdentity({
    userId: cfg.userId,
    deviceId,
    organizationId: cfg.organizationId,
  });
  agent.addServer({ publicKey: cfg.serverPublicKey });
  return {
    agent,
    dispose: () => {
      void agent.close();
    },
  };
}

/**
 * Drive the NHP-OTP step: ask nhp-server (via the relay) to email a
 * one-time code to `email`. The agent must have been built with the
 * same private key the backend sealed at registration.
 *
 * Returns {success} — the OTP itself is delivered out-of-band (email
 * or the nhp-server log fallback in docker).
 */
export async function requestOtp(
  handle: AgentHandle,
  cfg: NhpEndpointConfig,
  email: string,
): Promise<{ success: boolean; error?: string }> {
  console.log('[demoapp/nhp] requestOtp →', {
    relayUrl: cfg.relayUrl,
    serverPubKey: cfg.serverPublicKey.slice(0, 12) + '…',
    cipherScheme: cfg.cipherScheme,
    serviceId: cfg.serviceId,
    email,
  });
  const otpRes = await handle.agent.requestOtp(cfg.serviceId, { email });
  console.log('[demoapp/nhp] requestOtp result →', otpRes);
  return { success: otpRes.success, error: otpRes.error };
}

/**
 * Drive the NHP-REG step: submit the OTP + public key, get a RAK back.
 * The agent must have completed requestOtp under the same identity.
 */
export async function registerPublicKey(
  handle: AgentHandle,
  cfg: NhpEndpointConfig,
  email: string,
  otp: string,
): Promise<RegistrationResult> {
  console.log('[demoapp/nhp] registerPublicKey →', { serviceId: cfg.serviceId, otpLen: otp.length });
  const regRes = await handle.agent.registerPublicKey(cfg.serviceId, otp, { email });
  console.log('[demoapp/nhp] registerPublicKey result →', regRes);
  return {
    rakOk: regRes.success,
    expiresAt: regRes.expiresAt,
  };
}

/**
 * List resources the server says the agent can access.
 */
export async function listResources(
  handle: AgentHandle,
  cfg: NhpEndpointConfig,
): Promise<string[]> {
  const res = await handle.agent.listServices(cfg.serviceId);
  return res.success ? res.resources : [];
}

/**
 * Knock a single resource. Returns the first resolved host on success
 * (typically the AC's reverse-proxy host). Caller redirects to it.
 */
export async function knockResource(
  handle: AgentHandle,
  cfg: NhpEndpointConfig,
  resourceId: string,
): Promise<KnockOutcome> {
  const result = await handle.agent.knockResource({
    resourceId,
    serviceId: cfg.serviceId,
  });
  if (!result.success) {
    return { success: false, error: result.error || 'knock failed' };
  }
  const hosts = result.resourceHosts || {};
  const first = Object.values(hosts)[0];
  return { success: true, resourceHost: first };
}
