# Platform Integrations

This project should integrate with social media and marketplace platforms through official APIs or permitted user-driven workflows.

## Principles

- Respect platform terms, rate limits, permission scopes, and review requirements.
- Store inventory data in this app before publishing externally.
- Keep a record of what was published, when, where, and with which connected account.
- Make failures visible to the seller in plain language.
- Avoid behavior that resembles spam, ban evasion, credential sharing, or unauthorized scraping.

## Adapter Expectations

Each platform adapter should expose a small capability-oriented interface, such as:

- Validate account connection.
- Publish listing.
- Update listing.
- Mark listing unavailable where supported.
- Fetch publish status where supported.

Unsupported capabilities should be explicit so the frontend can communicate limitations clearly.

