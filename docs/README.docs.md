# Blockbusterr Documentation

This directory contains the Starlight documentation for Blockbusterr.

## Development

```bash
cd docs
npm install
npm run dev
```

Visit `http://localhost:4321/` to view the docs locally.

## Building

```bash
npm run build
```

Build the compatible `/v2/` path locally with:

```bash
DOCS_BASE=/v2/ npm run build
```

Built files will be in `dist/`.

## Deployment

Documentation is automatically deployed to GitHub Pages when documentation changes are pushed to `main`. The deployment builds the triggering commit:

- `/` serves stable v2 documentation.
- `/v2/` preserves existing v2 links with the same documentation.
- `/v1/` archives documentation from the immutable `v1.5.2` tag.

The archive build updates the version selector and prefixes legacy content links so navigation stays within `/v1/`.

View live docs at: https://blockbusterr.dev/

## Structure

```
src/content/docs/
├── index.mdx                    # Homepage
├── getting-started/
│   ├── quickstart.md
│   ├── installation.md
│   └── configuration.md
├── concepts/
│   ├── jobs.md
│   ├── rules.mdx
│   ├── smart-jobs.md
│   └── integration-modes.md
├── integrations/
│   ├── trakt.md
│   ├── radarr.md
│   ├── sonarr.md
│   ├── jellyseerr.md
│   └── tmdb.md
└── api/
    ├── overview.md
    ├── jobs.md
    ├── activity.md
    └── config.md
```

## Adding Pages

1. Create a new `.md` or `.mdx` file in `src/content/docs/`
2. Add frontmatter:
```md
---
title: Page Title
description: Page description
---
```
3. Update `astro.config.mjs` sidebar if needed
4. Content will be automatically included

## Learn More

- [Starlight Documentation](https://starlight.astro.build/)
- [Astro Documentation](https://docs.astro.build/)
