# Buffalo App Deployment to Coolify (Nixpacks)

This guide describes how to deploy the AVR NPO Buffalo-based site to Coolify using Nixpacks, including environment variable setup, database provisioning, migration automation, and post-deployment verification.

---

## Overview

- **Platform:** Coolify (self-hosted PaaS)
- **Build System:** Nixpacks (auto-detects Go/Buffalo)
- **App:** Buffalo (Go web framework)
- **Domain:** https://avrnpo.org
- **Database:** PostgreSQL (provisioned via Coolify)
- **Deployment Trigger:** Webhook on main branch merge

---

## Prerequisites

- Coolify instance running and accessible
- GitHub repo connected to Coolify
- Main branch webhook configured for auto-deploy
- Nixpacks selected as build pack in Coolify app settings

---

## Required Environment Variables

Set these in Coolify's app environment settings:

| Variable                | Purpose                        |
|-------------------------|--------------------------------|
| `GO_ENV`                | Set to `production`            |
| `SESSION_SECRET`        | Secure session secret (generate random string) |
| `HELCIM_PRIVATE_API_KEY`| Helcim payment API key         |
| `SMTP_HOST`             | Email SMTP server host         |
| `SMTP_PORT`             | Email SMTP server port         |
| `SMTP_USERNAME`         | Email SMTP username            |
| `SMTP_PASSWORD`         | Email SMTP password            |
| `FROM_EMAIL`            | Sender email address           |
| `FROM_NAME`            | Sender display name            |
| `DATABASE_URL`          | Postgres connection string     |
| `LOG_LEVEL` (optional)  | Logging level (`info`, `debug`)|
| `LOG_FILE_PATH` (optional)| Log file path                 |

**Note:**
- `DATABASE_URL` is provided by Coolify when you link the app to the Postgres service.
- Add any other secrets required by your app (see codebase for additional keys).

---

## Database Setup (Coolify)

1. **Provision Postgres:**
   - In Coolify dashboard, go to "Services" → "Add new service" → select **PostgreSQL**.
   - Configure as needed (name, version, credentials).
   - Start the service.

2. **Link App to Database:**
   - In your app's settings, link the Postgres service.
   - Coolify will expose an internal connection string (e.g., `postgres://user:pass@host:port/dbname`).
   - Set this as the `DATABASE_URL` environment variable for your app.

---

## App Configuration (Coolify/Nixpacks)

- **Build Command:** Nixpacks auto-detects Go/Buffalo and runs `go build` or `buffalo build`.
- **Start Command:** By default, Coolify will run the built binary (e.g., `./bin/app`).
- **Custom Build/Start:** If needed, add a `nixpacks.toml` to specify custom build/start commands:

```toml
[phases.build]
cmds = ["buffalo build -o bin/app"]

[phases.start]
cmds = ["./bin/app"]
```

---

## Automatic Migrations

To ensure database schema is up-to-date on every deploy:

- In Coolify app settings, add a **pre-start command**:
  - `soda migrate up`
- This runs migrations before starting the app.
- Alternatively, add to `nixpacks.toml`:

```toml
[phases.start]
cmds = ["soda migrate up", "./bin/app"]
```

---

## Deployment Flow

1. **Merge to main branch** → webhook triggers Coolify deploy
2. **Coolify pulls code** → runs Nixpacks build
3. **Runs migrations** (`soda migrate up`)
4. **Starts Buffalo app** (`./bin/app`)
5. **App is live at** https://avrnpo.org

---

## Verification Checklist

After deployment, verify:

- [ ] Site loads at https://avrnpo.org
- [ ] Email functions work (test with real/test credentials)
- [ ] Database is connected and migrations have run
- [ ] Static assets (CSS, JS, images) are served correctly
- [ ] No missing environment variables (check logs for errors)
- [ ] Admin dashboard and donation system function as expected

---

## Maintenance & Troubleshooting

- **Environment Variables:**
  - Update secrets in Coolify UI as needed
  - Check logs for missing/invalid env vars
- **Database:**
  - Use Coolify’s DB dashboard for backups, restores, and health checks
- **Migrations:**
  - Ensure `soda migrate up` runs on every deploy
  - Check migration status in logs if schema issues occur
- **Logs:**
  - Review logs in Coolify for errors, warnings, or missing config
- **Static Assets:**
  - Confirm `public/` directory is served by Buffalo
- **App Updates:**
  - On code changes, merge to main branch to trigger redeploy

---

## References

- [Coolify Nixpacks Build Pack](https://coolify.io/docs/builds/packs/nixpacks)
- [Coolify PostgreSQL Setup](https://coolify.io/docs/databases/postgresql)
- [Buffalo Framework Docs](https://gobuffalo.io/)
- [Nixpacks Customization](https://nixpacks.com/docs/guides/configuring-builds)

---

**For questions or issues, see the main README or contact the development team.**
