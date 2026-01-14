// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://blockbusterr.dev',
	base: '/',
	integrations: [
		starlight({
			title: 'Blockbusterr',
			description: 'Automate your media library with smart filters and scoring',
			tagline: 'Smart content discovery for your media server',
			
			// Logo (optional - add logo file to public/)
			// logo: {
			// 	src: './src/assets/logo.svg',
			// 	alt: 'Blockbusterr Logo',
			// },
			
			// Favicon (optional)
			favicon: '/favicon.svg',
			
			// Head tags for SEO
			head: [
				{
					tag: 'meta',
					attrs: {
						property: 'og:image',
						content: 'https://blockbusterr.dev/banner.png',
					},
				},
			],
			
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/mahcks/blockbusterr',
				},
			],
			
			// Edit link (optional - links to GitHub)
			editLink: {
				baseUrl: 'https://github.com/mahcks/blockbusterr/edit/master/docs/',
			},
			
			// Last updated timestamp
			lastUpdated: true,
			
			// Pagination (prev/next at bottom)
			pagination: true,
			
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'index' },
						{ label: 'Quick Start', slug: 'getting-started/quickstart' },
						{ label: 'Installation', slug: 'getting-started/installation' },
						{ label: 'Configuration', slug: 'getting-started/configuration' },
					],
				},
				{
					label: 'Core Concepts',
					items: [
						{ label: 'Jobs Overview', slug: 'concepts/jobs' },
						{ label: 'Filters & Scoring', slug: 'concepts/filters' },
						{ label: 'Smart Jobs', slug: 'concepts/smart-jobs' },
						{ label: 'Integration Modes', slug: 'concepts/integration-modes' },
					],
				},
				{
					label: 'Examples',
					items: [
						{ label: 'Use Cases & Examples', slug: 'examples/use-cases' },
					],
				},
				{
					label: 'Integrations',
					items: [
						{ label: 'Trakt', slug: 'integrations/trakt' },
						{ label: 'Radarr', slug: 'integrations/radarr' },
						{ label: 'Sonarr', slug: 'integrations/sonarr' },
						{ label: 'Jellyseerr', slug: 'integrations/jellyseerr' },
						{ label: 'TMDB (Optional)', slug: 'integrations/tmdb' },
					],
				},
				{
					label: 'API Reference',
					items: [
						{ label: 'Overview', slug: 'api/overview' },
						{ label: 'Jobs API', slug: 'api/jobs' },
						{ label: 'Activity API', slug: 'api/activity' },
						{ label: 'Configuration API', slug: 'api/config' },
					],
				},
			],
			customCss: ['./src/styles/custom.css'],
		}),
	],
});
