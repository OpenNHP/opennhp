// encryptPrivateKey / decryptPrivateKey for the Demo App.
//
// The NHP private key lives in the browser only. We encrypt it with a
// key derived from the user's password via PBKDF2-SHA256 (200k
// iterations) and AES-256-GCM. The wire format is:
//
//   base64( salt[16] | iv[12] | ciphertext )
//
// We intentionally keep salt+iv with the blob so the format is
// self-describing: a later code path only needs (encryptedBlob, password)
// to decrypt, with no other parameters required.
const PBKDF2_ITERATIONS = 200_000;
const SALT_BYTES = 16;
const IV_BYTES = 12;

export async function encryptPrivateKey(privKeyBase64, password) {
    if (!privKeyBase64 || !password) {
        throw new Error('encryptPrivateKey: missing input');
    }
    const salt = crypto.getRandomValues(new Uint8Array(SALT_BYTES));
    const iv = crypto.getRandomValues(new Uint8Array(IV_BYTES));
    const aesKey = await deriveAesKey(password, salt);
    const plain = base64ToBytes(privKeyBase64);
    const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, aesKey, plain);
    return bytesToBase64(concatBytes(salt, iv, new Uint8Array(ct)));
}

export async function decryptPrivateKey(encryptedBase64, password) {
    if (!encryptedBase64 || !password) {
        throw new Error('decryptPrivateKey: missing input');
    }
    const blob = base64ToBytes(encryptedBase64);
    if (blob.length < SALT_BYTES + IV_BYTES + 16) {
        throw new Error('decryptPrivateKey: blob too short');
    }
    const salt = blob.slice(0, SALT_BYTES);
    const iv = blob.slice(SALT_BYTES, SALT_BYTES + IV_BYTES);
    const ct = blob.slice(SALT_BYTES + IV_BYTES);
    const aesKey = await deriveAesKey(password, salt);
    let plain;
    try {
        plain = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, aesKey, ct);
    } catch (e) {
        throw new Error('decryptPrivateKey: wrong password or corrupted blob');
    }
    return bytesToBase64(new Uint8Array(plain));
}

async function deriveAesKey(password, salt) {
    const km = await crypto.subtle.importKey(
        'raw', new TextEncoder().encode(password),
        'PBKDF2', false, ['deriveKey'],
    );
    return crypto.subtle.deriveKey(
        { name: 'PBKDF2', salt, iterations: PBKDF2_ITERATIONS, hash: 'SHA-256' },
        km,
        { name: 'AES-GCM', length: 256 },
        false, ['encrypt', 'decrypt'],
    );
}

function concatBytes(...arrays) {
    const total = arrays.reduce((n, a) => n + a.length, 0);
    const out = new Uint8Array(total);
    let off = 0;
    for (const a of arrays) { out.set(a, off); off += a.length; }
    return out;
}

function bytesToBase64(bytes) {
    let s = '';
    for (let i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]);
    return btoa(s);
}

function base64ToBytes(b64) {
    const s = atob(b64);
    const out = new Uint8Array(s.length);
    for (let i = 0; i < s.length; i++) out[i] = s.charCodeAt(i);
    return out;
}
