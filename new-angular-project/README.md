# Marketino Admin (new-ui)

پنل مدیریت مارکتینو — Angular 21، طراحی راست‌به‌چپ (فارسی).

## Development

```bash
npm install
npm start        # serves on http://localhost:4201 (proxies /api → http://localhost:3000)
```

The Go backend (`../marketino`) runs on `:3000` and reverse-proxies non-API routes
to this dev server on `:4201`, so either origin works during development.

## Structure

- `src/app/core/auth` — token storage, HTTP interceptor, route guards
- `src/app/core/api` — dataview + entity services (users, roles)
- `src/app/shared` — data table, confirm dialog, toasts, icons
- `src/app/features` — login, shell (collapsible sidebar), dashboard, users, roles

## Notes

- The npm registry mirror is missing some package tarballs; `package.json`
  contains `overrides` pinning affected transitive dependencies to versions
  the mirror can actually serve.
- The backend has no get-by-id endpoint, so record pages opened by direct URL
  are resolved by scanning the dataview pages (`DataviewService.findById`).
