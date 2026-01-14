# Blockbusterr Documentation

This directory contains the Starlight documentation for Blockbusterr.

## Development

```bash
cd docs
npm install
npm run dev
```

Visit `http://localhost:4321/blockbusterr/` to view the docs locally.

## Building

```bash
npm run build
```

Built files will be in `dist/`.

## Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch.

View live docs at: https://mahcks.github.io/blockbusterr/

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
│   ├── filters.md
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
