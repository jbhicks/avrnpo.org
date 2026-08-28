# American Veterans Rebuilding (AVR NPO)

[![test](https://github.com/jbhicks/avrnpo.org/actions/workflows/test.yml/badge.svg)](https://github.com/jbhicks/avrnpo.org/actions/workflows/test.yml)

Go/PocketBase site for [American Veterans Rebuilding](https://avrnpo.org), a 501(c)(3) helping combat veterans through housing, skills training, and community programs. Live at **https://avrnpo.org**.

**Stack:** Go 1.24, PocketBase, Templ, HTMX, Pico CSS, Helcim donations, SQLite.

## About AVR NPO

American Veterans Rebuilding is formed by combat veterans of the wars in Afghanistan and Iraq. The organization is dedicated to the military core values of Loyalty, Duty, Respect, Selfless Service, Honor, Integrity, and Personal Courage.

## Quick Start

Prerequisites: Go 1.24+ and the Templ CLI (`go install github.com/a-h/templ/cmd/templ@latest`).

```console
cp .env.example .env   # set PB_ADMIN_EMAIL, PB_ADMIN_PASSWORD, SMTP, Helcim
make install           # Air + Templ + modules
make dev               # http://127.0.0.1:8090
```

Admin UI: http://127.0.0.1:8090/_/

```console
go test ./...                    # unit tests
E2E_TESTS=1 go test -v -run E2E  # browser e2e
make build                       # production binary
```

## Features

- Mission, team, projects, contact
- Helcim donations (one-time and monthly) with receipts
- Blog / CMS via PocketBase admin
- HTMX navigation, SEO meta tags

## Helcim

Uses Helcim Pay (`https://secure.helcim.app/helcim-pay/services/start.js`, `POST /v2/helcim-pay/initialize`). Local donate page: http://127.0.0.1:8090/donate. Test card `4124939999999990` (CVV 100). See [docs/payment-system](./docs/payment-system/).

## Documentation

- [Quick start](./docs/getting-started/quick-start.md)
- [Development workflow](./docs/getting-started/development-workflow.md)
- [Payment system](./docs/payment-system/README.md)
- [Coolify deployment](./docs/deployment/coolify-pocketbase-migration.md)
- [Security](./docs/deployment/security.md)

## Contributing

Public source for AVR NPO. Review `docs/getting-started/`, run `go test ./...`, and follow [security guidelines](./docs/deployment/security.md).

## Contact

- Site: [avrnpo.org](https://avrnpo.org)
- Programs: michael@avrnpo.org

Content and imagery related to American Veterans Rebuilding is proprietary to the organization. Website code is built on open-source technologies.
