// register.js — three-step registration wizard.
//
// 1. /api/users/register: create the Demo user (server stores the
//    encrypted private key).
// 2. POST NHP_OTP via js-agent + relay (fire-and-forget; the demo-app
//    plugin emails the OTP code).
// 3. POST NHP_REG via js-agent + relay; success means the nhp-server
//    has accepted the public key. The plugin will then call
//    /internal/nhp/mark-registered which flips our user row to "done".
//
// Throughout, we keep `agent`, `password`, and `pubKey` in module
// scope; the user's private key never leaves memory in plaintext
// beyond what the agent needs internally.

import { NHPAgent } from '/static/nhp-agent.esm.js';
import { encryptPrivateKey } from '/static/crypto.js';

const { authServiceId, cipherScheme } = window.__NHP__ || {};

let agent = null;
let password = '';
let username = '';
let pubKey = '';

const stepAccount = document.getElementById('step-account');
const stepOtp = document.getElementById('step-otp');
const stepDone = document.getElementById('step-done');
const step1Err = document.getElementById('step1-error');
const step2Err = document.getElementById('step2-error');

document.getElementById('register-form').addEventListener('submit', onRegisterSubmit);
document.getElementById('otp-form').addEventListener('submit', onOtpSubmit);

async function onRegisterSubmit(e) {
    e.preventDefault();
    step1Err.hidden = true;

    const fd = new FormData(e.target);
    username = fd.get('username').toString().trim();
    password = fd.get('password').toString();
    const email = fd.get('email').toString().trim();

    try {
        // 1. Generate NHP keypair in the browser. The private key stays
        //    in memory; we encrypt it before sending the public key up.
        agent = new NHPAgent({ cipherScheme });
        await agent.init();
        pubKey = agent.getPublicKey();
        const privKey = agent.getPrivateKey();
        const encryptedPrivateKey = await encryptPrivateKey(privKey, password);

        // 2. Create the Demo user with the encrypted private key.
        const url = '/api/nhp/config'; // authenticated; uses session if any
        const cfgRes = await fetch(url, { credentials: 'include' });
        const nhpConfig = await cfgRes.json();

        agent.setIdentity({ userId: username, deviceId: 'browser-reg', organizationId: '' });
        agent.addServer({
            publicKey: nhpConfig.serverPubKey,
            relayUrl: nhpConfig.relayUrl,
            id: 'nhp-server',
        });

        const regRes = await fetch('/api/users/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                username, password, email, encryptedPrivateKey, nhpPublicKey: pubKey,
            }),
        });
        if (!regRes.ok) {
            const j = await regRes.json().catch(() => ({}));
            throw new Error(j.error || `register failed: ${regRes.status}`);
        }

        // 3. Fire OTP through the relay. The plugin will pick up the
        //    email from UserData and call /internal/nhp/otp-deliver.
        await agent.requestOtp(authServiceId, { email });

        // 4. Move to step 2.
        showStep(stepOtp);
    } catch (err) {
        console.error(err);
        step1Err.textContent = err.message || String(err);
        step1Err.hidden = false;
    }
}

async function onOtpSubmit(e) {
    e.preventDefault();
    step2Err.hidden = true;
    const fd = new FormData(e.target);
    const otp = fd.get('otp').toString().trim();

    try {
        // Subscribe the same identity / server config the agent already
        // has. We refresh NHP config to pick up any rotation between
        // step 1 and step 2.
        const cfg = await fetch('/api/nhp/config').then(r => r.json());
        agent.setIdentity({ userId: username, deviceId: 'browser-reg', organizationId: '' });
        agent.addServer({
            publicKey: cfg.serverPubKey,
            relayUrl: cfg.relayUrl,
            id: 'nhp-server',
        });

        const result = await agent.registerPublicKey(authServiceId, otp);
        if (!result.success) {
            throw new Error(result.error || 'Registration failed');
        }
        showStep(stepDone);
    } catch (err) {
        console.error(err);
        step2Err.textContent = err.message || String(err);
        step2Err.hidden = false;
    }
}

function showStep(el) {
    for (const s of [stepAccount, stepOtp, stepDone]) s.classList.remove('active');
    el.classList.add('active');
}
