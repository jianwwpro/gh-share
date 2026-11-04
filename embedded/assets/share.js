/*
 * share.js — optional progressive enhancement. Pages work without it.
 *   - copy button on every <pre>
 *   - syntax highlighting (highlight.js from CDN, loaded only when a
 *     <code class="language-*"> exists; colors come from --syn-* in share.css)
 * Include with: <script src="../../assets/share.js" defer></script>
 */
(() => {
  const HLJS = "https://cdn.jsdelivr.net/npm/@highlightjs/cdn-assets@11.11.1/highlight.min.js";

  function addCopyButtons() {
    document.querySelectorAll("pre").forEach((pre) => {
      if (pre.querySelector(".copy-btn")) return;
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "copy-btn";
      btn.textContent = "Copy";
      btn.addEventListener("click", async () => {
        const code = pre.querySelector("code") || pre;
        try {
          await navigator.clipboard.writeText(code.innerText.replace(/\n$/, ""));
          btn.textContent = "Copied";
        } catch {
          btn.textContent = "Failed";
        }
        setTimeout(() => (btn.textContent = "Copy"), 1500);
      });
      pre.appendChild(btn);
    });
  }

  function highlight() {
    const blocks = document.querySelectorAll('pre code[class*="language-"]');
    if (!blocks.length) return;
    const s = document.createElement("script");
    s.src = HLJS;
    s.onload = () => blocks.forEach((el) => window.hljs.highlightElement(el));
    document.head.appendChild(s);
  }

  function init() {
    addCopyButtons();
    highlight();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
