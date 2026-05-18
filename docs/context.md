# Project Context

## Purpose

Build a private social media marketplace assistant for a small resale workflow. The app should let non-technical users capture inventory details once, keep those records in a durable source of truth, and publish unsold inventory to connected social media or marketplace accounts.

The project is not intended for malicious automation, spam, account evasion, or platform abuse. It should prefer official APIs and permitted workflows. If a platform account becomes unavailable, the business value is the retained inventory record and the ability to connect another supported account where permitted.

## Primary Users

- Seller: non-technical user who captures item photos, descriptions, prices, condition, size, and listing status.
- Partner: collaborator who can help manage inventory and listings.
- Developer/operator: technical maintainer responsible for integrations, deployment, data safety, and support.

## Product Goals

- Capture goods such as clothing, shoes, accessories, and small household items.
- Store item details, media, listing state, sale state, and platform publishing history centrally.
- Support publishing to multiple connected social or marketplace accounts over time.
- Keep unsold inventory easy to find, edit, relist, or migrate.
- Provide a simple, reliable interface optimized for a non-technical seller.

## Non-Goals

- Bypassing platform rate limits, bans, verification, anti-abuse systems, or terms of service.
- Scraping or automating platforms where the workflow is disallowed.
- Building a generic public SaaS product before the private workflow is proven.

## Initial Technical Choices

- Monorepo.
- Frontend in Angular and TypeScript under `apps/web`.
- Backend in Go under `services/api`.
- Documentation under `docs`.
- Start with a small API health check and a frontend shell so the app can deploy early.

## Open Decisions

- Database choice and hosting model.
- Object storage for item photos.
- Authentication and authorization model.
- First supported social or marketplace integration.
- Deployment target and infrastructure-as-code approach.

## Assistant Guidance

Future assistant sessions should read this file first, then inspect the current repository state before making changes. Preserve the platform-compliant automation boundary when designing features or integrations.
