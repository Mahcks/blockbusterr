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

Build the v2 beta path locally with:

```bash
DOCS_BASE=/v2/ npm run build
```

Built files will be in `dist/`.

## Deployment

Documentation is automatically deployed to GitHub Pages when documentation changes are pushed to `main` or `release/v2.0.0`. The deployment combines both branches:

- `/` serves the stable v1 documentation from `main`.
- `/v2/` serves the beta documentation from `release/v2.0.0`.

When v2 becomes stable, archive the final v1 documentation under `/v1/`, make
`main` the unversioned v2 source, and redirect `/v2/` to `/`.

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
