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
import { t } from './i18n.js';
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
  // logLevel defaults to 'error' (silent for info / debug). Production
  // never opts into debug; the demo console was previously drowned in
  // requestOtp / registerPublicKey noise that included the user's email
  // + relay URL + serviceId + full OTP result objects (review follow-up
  // to #6).
  const logLevel = opts.logLevel ?? 'error';
  const debug = logLevel === 'debug';

  function renderWaiting(message: string): void {
    container.innerHTML = `
      <div class="panel">
        <h2>${t('nhp.title')}</h2>
        <div class="spinner"></div>
        <p class="note">${escapeHtml(message)}</p>
        ${opts.onBack ? `<button class="btn btn-secondary" id="nhp-back">${t('nhp.back')}</button>` : ''}
      </div>
    `;
    container.querySelector<HTMLButtonElement>('#nhp-back')?.addEventListener('click', () => opts.onBack?.());
  }

  function renderOtpEntry(): void {
    container.innerHTML = `
      <div class="panel">
        <h2>${t('nhp.title')}</h2>
        <p class="note">${t('nhp.otpSentTo', { email: escapeHtml(opts.email) })}</p>
        <div id="nhp-alert"></div>
        <div class="field">
          <label for="nhp-otp">${t('nhp.otpLabel')}</label>
          <input id="nhp-otp" type="text" autocomplete="one-time-code" inputmode="numeric"
                 placeholder="000000" maxlength="6" />
        </div>
        <button id="nhp-register" class="btn btn-primary">${t('nhp.registerBtn')}</button>
        <button id="nhp-resend" class="btn btn-secondary">${t('nhp.resend')}</button>
        ${opts.onBack ? `<button class="btn btn-secondary" id="nhp-back">${t('nhp.back')}</button>` : ''}
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
      resendBtn.textContent = t('nhp.resendCooldown', { n: remaining });
      resendTimer = setInterval(() => {
        remaining -= 1;
        if (remaining <= 0) {
          if (resendTimer) clearInterval(resendTimer);
          resendTimer = null;
          resendBtn.disabled = false;
          resendBtn.textContent = t('nhp.resend');
        } else {
          resendBtn.textContent = t('nhp.resendCooldown', { n: remaining });
        }
      }, 1000);
    };
    resendBtn.addEventListener('click', async () => {
      if (resendBtn.disabled) return;
      if (!handle) {
        showAlert('error', t('nhp.agentNotReady'));
        return;
      }
      resendBtn.disabled = true;
      showAlert('info', t('nhp.requestingNew'));
      try {
        const otp = await requestOtp(handle, opts.nhp, opts.email, debug);
        if (otp.success) {
          showAlert('info', t('nhp.newOtpSent', { email: opts.email }));
        } else {
          showAlert('error', t('nhp.otpReqFailed', { suffix: otp.error ? ': ' + otp.error : '' }));
        }
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        showAlert('error', t('nhp.otpReqFailed', { suffix: ': ' + msg }));
      } finally {
        // Apply the cooldown whether the request succeeded or failed so
        // a caller cannot bypass the throttle by inducing errors.
        startResendCooldown();
      }
    });

    registerBtn.addEventListener('click', async () => {
      const otp = otpInput.value.trim();
      if (!otp) {
        showAlert('error', t('nhp.enterOtp'));
        return;
      }
      if (!handle) {
        showAlert('error', t('nhp.agentNotReady'));
        return;
      }
      registerBtn.disabled = true;
      showAlert('info', t('nhp.driving'));
      try {
        const result = await registerPublicKey(handle, opts.nhp, opts.email, otp, debug);
        if (!result.rakOk) {
          showAlert('error', t('nhp.regFailedCheck'));
          return;
        }
        await api.registerConfirm(opts.regToken ?? '', opts.deviceId, result.expiresAt ?? 0, true);
        showAlert('success', t('nhp.complete'));
        opts.onComplete(true);
      } catch (err) {
        const msg = err instanceof ApiError || err instanceof Error ? err.message : String(err);
        showAlert('error', t('nhp.handshakeFailed', { msg }));
      } finally {
        registerBtn.disabled = false;
      }
    });
  }

  function renderError(message: string): void {
    container.innerHTML = `
      <div class="panel">
        <h2>${t('nhp.title')}</h2>
        <div class="alert alert-error">${escapeHtml(message)}</div>
        ${opts.onBack ? `<button class="btn btn-secondary" id="nhp-back">${t('nhp.back')}</button>` : ''}
      </div>
    `;
    container.querySelector<HTMLButtonElement>('#nhp-back')?.addEventListener('click', () => opts.onBack?.());
  }

  // Kick off requestOtp on mount.
  (async () => {
    renderWaiting(t('nhp.creatingAgent'));
    try {
      handle = await createAgent(opts.privateKey, opts.nhp, opts.deviceId, logLevel);
      const otp = await requestOtp(handle, opts.nhp, opts.email, debug);
      if (disposed) return;
      if (!otp.success) {
        renderError(t('nhp.otpReqFailed', { suffix: otp.error ? ': ' + otp.error : '' }));
        return;
      }
      renderOtpEntry();
    } catch (err) {
      if (disposed) return;
      const msg = err instanceof Error ? err.message : String(err);
      renderError(t('nhp.otpReqFailed', { suffix: ': ' + msg }));
    }
  })();

  // Dispose function: idempotent.
  return () => {
    disposed = true;
    handle?.dispose();
    handle = undefined;
  };
}
