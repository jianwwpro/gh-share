---
name: share
description: Publish content as a public web page at yuler.github.io/share and return the link — an AI/agent conversation, an idea, a code snippet, images or video. Use when the user asks to share something or wants a public link for it.
---

# share

Turns content into a public static web page on GitHub Pages (`yuler/share`) and returns its public URL.

## Recommended: `gh share post`

The fastest way to share content is using the `gh share` CLI:

- **Share a file or snippet**:
  ```bash
  gh share post path/to/snippet.ts --title "Debounce helper" --lede "Short explanation"
  ```
- **Share an AI conversation or note**:
  ```bash
  gh share post chat.html --type chat --title "Architecture discussion"
  ```
- **Share from standard input**:
  ```bash
  echo "# My Idea\n\nContent..." | gh share post --title "Product Idea" --type note
  ```
- **List existing shares**:
  ```bash
  gh share list
  ```

## Direct repo workflow

If you need full control over custom HTML or multiple bundled assets (e.g. images next to `index.html`):
1. **Repo**: Work in `~/Projects/share`. If missing: `gh repo clone yuler/share ~/Projects/share`.
2. **Conventions**: Consult `~/Projects/share/AGENTS.md` for slug naming (`YYYY-MM-DD-<kebab-title>`), template borrowing, asset placement, and theme options (`paper`, `dark`, `terminal`).
3. **Redact**: Strip tokens, keys, passwords, private emails, phone numbers, and local paths (`/home/...`).
4. **Push**:
   ```bash
   git add s/<slug> && git commit -m "📝 share: <title>" && git push
   ```

## Completion

Reply with the public URL (`https://yuler.github.io/share/s/<slug>/`) and mention that GitHub Pages updates within ~1 minute.
