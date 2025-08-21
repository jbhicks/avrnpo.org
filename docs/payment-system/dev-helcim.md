Development safety for Helcim integration

This document explains how to avoid creating real payments or subscriptions when developing locally.

Recommended local setup

- Set GO_ENV=development in your local environment (this repo's code checks this value).
- Do NOT set HELCIM_PRIVATE_API_KEY in your local shell unless you intentionally want real Helcim API calls.
  - If you must set a key for development, prefer a Helcim sandbox/test key (if provided by Helcim). Do not use production keys in development or CI.
- Optionally set HELCIM_WEBHOOK_VERIFIER_TOKEN in staging/production only. The code bypasses webhook signature verification when GO_ENV=="development".

What this repo does to keep dev safe

- One-time payments: the code simulates successful payments when GO_ENV=="development" (no Helcim calls).
- Recurring subscriptions: the code now simulates plan and subscription creation in development (no Helcim calls) and updates the DB with simulated IDs.
- NewHelcimClient: in development, if HELCIM_PRIVATE_API_KEY is not set, the client will be constructed without panicking, preventing crashes while still avoiding real API calls.

Testing webhooks locally

- Use a local tunneling tool (ngrok or similar) to expose your app to Helcim's webhook tester or to replay webhook requests.
- In development you can also craft local webhook POSTs; the app will skip signature verification when GO_ENV=="development".

Checklist before running anything that touches payments

- [ ] Confirm GO_ENV=development (local shell)
- [ ] Confirm HELCIM_PRIVATE_API_KEY is unset or set to sandbox/test key
- [ ] Confirm HELCIM_WEBHOOK_VERIFIER_TOKEN is unset in local environment

If you want, I can add a short test that asserts NewHelcimClient does not panic when HELCIM_PRIVATE_API_KEY is missing and GO_ENV=development.
