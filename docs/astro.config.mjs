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
			description: 'Automate media discovery with reusable rules and observable delivery',
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
				{
					tag: 'script',
					attrs: {
						defer: true,
						src: 'https://static.cloudflareinsights.com/beacon.min.js',
						'data-cf-beacon': '{"token": "d3a46d82612243b18b3bc3e640f37803"}',
					},
				},
			],
			
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/mahcks/blockbusterr',
				},
				{
					icon: 'discord',
					label: 'Discord',
					href: 'https://discord.com/invite/c8vb3VZqmg',
				}
			],
			
			// Edit link (optional - links to GitHub)
			editLink: {
				baseUrl: 'https://github.com/mahcks/blockbusterr/edit/main/docs/',
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
						{ label: 'Upgrading to v2', slug: 'getting-started/upgrading-to-v2' },
					],
				},
				{
					label: 'Core Concepts',
					items: [
						{ label: 'Jobs Overview', slug: 'concepts/jobs' },
						{ label: 'Rules', slug: 'concepts/filters' },
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
						{ label: 'Simkl (Optional)', slug: 'integrations/simkl' },
					],
				},
				{
					label: 'API Reference',
					items: [
						{ label: 'Overview', slug: 'api/overview' },
						{ label: 'Jobs API', slug: 'api/jobs' },
						{ label: 'Rules API', slug: 'api/rules' },
						{ label: 'Activity API', slug: 'api/activity' },
						{ label: 'Configuration API', slug: 'api/config' },
					],
				},
			],
			customCss: ['./src/styles/custom.css'],
		}),
	],
});
