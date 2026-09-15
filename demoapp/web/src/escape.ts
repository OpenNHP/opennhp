// HTML-escape a string for safe interpolation into template literals
// rendered via innerHTML. Used in both element-text contexts
// (`<span>${escape(s)}</span>`) and double-quoted attribute contexts
// (`data-foo="${escape(s)}"`, `<option value="${escape(s)}">`).
//
// We escape & first so the &-introduced entities we emit (&amp; / &lt;
// / &gt; / &quot; / &#39;) are not double-escaped. The five characters
// are the minimum set that closes every HTML / attribute-injection
// gap; double and single quotes matter because the helper is used
// inside data-* attributes (e.g. resources.ts data-resid / data-url)
// and <option value="...">. Inline <script> is blocked by the demoapp's
// CSP (`script-src 'self'`), so even a quote-break is not currently
// exploitable, but a single misconfigured CSP or future relaxed rule
// would make this trivially exploitable — and the fix is one line.
//
// Kept as a shared module (rather than a per-view copy) so every
// caller gets the same character set without each view needing to
// remember to add quote handling.
//
// Originally this helper lived as a private `function escape` inside
// login.ts / register.ts / complete-registration.ts / resources.ts /
// nhp-reg-panel.ts, all copies identical and all missing the quote
// handling. A resource Id/URL/Title or server Name with a quote broke
// out of the attribute.

export function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}