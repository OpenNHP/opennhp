// Shared NHP registration panel — owns the requestOtp → registerPublicKey →
// confirm state machine so both the fresh-registration view (register.ts)
// and the resume view (complete-registration.ts) render the same UX.
//
// On mount it builds an NHPAgent from the backend-issued private key and
// immediately fires requestOtp (NHP-OTP via the relay → nhp-server emails
// the code). The user then enters the code and clicks "Register public
// key", which sends NHP-REG and, on a successful RAK, calls
// /api/register/confirm to flip the account to active.

import { api, ApiError, type NhpEndpointConfig } from './api.js';
import { escapeHtml } from './escape.js';
import { createAgent, requestOtp, registerPublicKey, type AgentHandle } from './nhp.js';

export interface NhpRegPanelOpts {
  privateKey: string;
  deviceId: string;
  email: string;
  nhp: NhpEndpointConfig;
  /** Present for fresh registration; omitted for the resume (session) path. */
  regToken?: string;
  logLevel?: 'silent' | 'error' | 'info' | 'debug';
  /** Called after /api/register/confirm resolves; rakOk reflects the RAK. */
  onComplete: (rakOk: boolean) => void;
  /** Optional "back" / cancel — shown while waiting for the OTP. */
  onBack?: () => void;
}

/**
 * Mount the panel into `container`. Returns a dispose function that closes
 * the underlying agent (wiping the in-memory private key reference) — call
 * it on unmount or after completion.
 */
export function mountNhpRegPanel(container: HTMLElement, opts: NhpRegPanelOpts): () => void {
  let handle: AgentHandle | undefined;
  let disposed = false;

  function renderWaiting(message: string): void {
    container.innerHTML = `
      <div class="panel">
        <h2>NHP registration</h2>
        <div class="spinner"></div>
        <p class="note">${escapeHtml(message)}</p>
        ${opts.onBack ? '<button class="btn btn-secondary" id="nhp-back">Back</button>' : ''}
      </div>
    `;
    container.querySelector<HTMLButtonElement>('#nhp-back')?.addEventListener('click', () => opts.onBack?.());
  }

  function renderOtpEntry(): void {
    container.innerHTML = `
      <div class="panel">
        <h2>NHP registration</h2>
        <p class="note">An OTP has been sent to <strong>${escapeHtml(opts.email)}</strong>.
          In docker environments, check the nhp-server logs for the OTP fallback.</p>
        <div id="nhp-alert"></div>
        <div class="field">
          <label for="nhp-otp">OTP</label>
          <input id="nhp-otp" type="text" autocomplete="one-time-code" inputmode="numeric"
                 placeholder="000000" maxlength="6" />
        </div>
        <button id="nhp-register" class="btn btn-primary">Register public key with nhp-server</button>
        <button id="nhp-resend" class="btn btn-secondary">Resend code</button>
        ${opts.onBack ? '<button class="btn btn-secondary" id="nhp-back">Back</button>' : ''}
      </div>
    `;
    const alert = container.querySelector<HTMLDivElement>('#nhp-alert')!;
    const otpInput = container.querySelector<HTMLInputElement>('#nhp-otp')!;
    const registerBtn = container.querySelector<HTMLButtonElement>('#nhp-register')!;
    const resendBtn = container.querySelector<HTMLButtonElement>('#nhp-resend')!;

    const showAlert = (level: 'error' | 'info' | 'success', msg: string) => {
      alert.innerHTML = `<div class="alert alert-${level}">${escapeHtml(msg)}</div>`;
    };

    container.querySelector<HTMLButtonElement>('#nhp-back')?.addEventListener('click', () => opts.onBack?.());

    // Resend re-fires requestOtp on the existing agent — no account
    // re-creation, so it avoids the 409 you'd hit by clicking "Create
    // account" again on the fresh-registration view.
    //
    // Cooldown: the OTP is sent directly browser → relay → nhp-server
    // (NOT through the demoapp, so the server-side /api/register rate
    // limit does not cover it). To blunt the email-amplification
    // primitive flagged in review #6 we disable Resend for a cooldown
    // after each click. This is client-side friction only; a hard
    // per-IP/per-email OTP-send limit belongs in nhp-server's basic
    // plugin (tracked as a follow-up).
    const RESEND_COOLDOWN_SECONDS = 30;
    let resendTimer: ReturnType<typeof setInterval> | null = null;
    const startResendCooldown = () => {
      if (resendTimer) clearInterval(resendTimer);
      let remaining = RESEND_COOLDOWN_SECONDS;
      resendBtn.disabled = true;
      resendBtn.textContent = `Resend code (${remaining}s)`;
      resendTimer = setInterval(() => {
        remaining -= 1;
        if (remaining <= 0) {
          if (resendTimer) clearInterval(resendTimer);
          resendTimer = null;
          resendBtn.disabled = false;
          resendBtn.textContent = 'Resend code';
        } else {
          resendBtn.textContent = `Resend code (${remaining}s)`;
        }
      }, 1000);
    };
    resendBtn.addEventListener('click', async () => {
      if (resendBtn.disabled) return;
      if (!handle) {
        showAlert('error', 'Agent not ready. Go back and try again.');
        return;
      }
      resendBtn.disabled = true;
      showAlert('info', 'Requesting a new OTP…');
      try {
        const otp = await requestOtp(handle, opts.nhp, opts.email);
        if (otp.success) {
          showAlert('info', `A new OTP has been sent to ${opts.email}.`);
        } else {
          showAlert('error', `NHP-OTP request failed${otp.error ? ': ' + otp.error : ''}.`);
        }
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        showAlert('error', `NHP-OTP request failed: ${msg}`);
      } finally {
        // Apply the cooldown whether the request succeeded or failed so
        // a caller cannot bypass the throttle by inducing errors.
        startResendCooldown();
      }
    });

    registerBtn.addEventListener('click', async () => {
      const otp = otpInput.value.trim();
      if (!otp) {
        showAlert('error', 'Enter the OTP from your email.');
        return;
      }
      if (!handle) {
        showAlert('error', 'Agent not ready. Go back and try again.');
        return;
      }
      registerBtn.disabled = true;
      showAlert('info', 'Driving NHP-REG handshake…');
      try {
        const result = await registerPublicKey(handle, opts.nhp, opts.email, otp);
        if (!result.rakOk) {
          showAlert('error', 'NHP registration failed — check the OTP and try again.');
          return;
        }
        await api.registerConfirm(opts.regToken ?? '', opts.deviceId, result.expiresAt ?? 0, true);
        showAlert('success', 'Registration complete.');
        opts.onComplete(true);
      } catch (err) {
        const msg = err instanceof ApiError || err instanceof Error ? err.message : String(err);
        showAlert('error', `NHP handshake failed: ${msg}`);
      } finally {
        registerBtn.disabled = false;
      }
    });
  }

  function renderError(message: string): void {
    container.innerHTML = `
      <div class="panel">
        <h2>NHP registration</h2>
        <div class="alert alert-error">${escapeHtml(message)}</div>
        ${opts.onBack ? '<button class="btn btn-secondary" id="nhp-back">Back</button>' : ''}
      </div>
    `;
    container.querySelector<HTMLButtonElement>('#nhp-back')?.addEventListener('click', () => opts.onBack?.());
  }

  // Kick off requestOtp on mount.
  (async () => {
    renderWaiting('Creating NHP agent and requesting OTP…');
    try {
      handle = await createAgent(opts.privateKey, opts.nhp, opts.deviceId, opts.logLevel ?? 'debug');
      const otp = await requestOtp(handle, opts.nhp, opts.email);
      if (disposed) return;
      if (!otp.success) {
        renderError(`NHP-OTP request failed${otp.error ? ': ' + otp.error : ''}.`);
        return;
      }
      renderOtpEntry();
    } catch (err) {
      if (disposed) return;
      const msg = err instanceof Error ? err.message : String(err);
      renderError(`NHP-OTP request failed: ${msg}`);
    }
  })();

  // Dispose function: idempotent.
  return () => {
    disposed = true;
    handle?.dispose();
    handle = undefined;
  };
}
