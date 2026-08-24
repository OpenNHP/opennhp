// login.js — POST credentials, redirect on success.
const form = document.getElementById('login-form');
const errorEl = document.getElementById('login-error');

form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorEl.hidden = true;
    const fd = new FormData(form);
    const res = await fetch('/api/users/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            username: fd.get('username'),
            password: fd.get('password'),
        }),
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
        errorEl.textContent = data.error || 'Login failed';
        errorEl.hidden = false;
        return;
    }
    const next = new URLSearchParams(location.search).get('next') || '/dashboard';
    location.href = next;
});
