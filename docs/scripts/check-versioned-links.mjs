import fs from 'node:fs';
import path from 'node:path';

const root = process.argv[2] || 'dist';
const prefixes = ['/getting-started/', '/concepts/', '/integrations/', '/api/', '/examples/'];
const invalid = [];
const origin = 'https://blockbusterr.dev';

function deployedPath(file) {
	const relative = path.relative(root, file).replaceAll(path.sep, '/');
	return `/v2/${relative === 'index.html' ? '' : relative.replace(/index\.html$/, '')}`;
}

function targetFile(pathname) {
	const relative = pathname.slice('/v2/'.length);
	return path.join(root, relative.endsWith('/') || relative === '' ? relative + 'index.html' : relative);
}

function scan(directory) {
	for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
		const file = path.join(directory, entry.name);
		if (entry.isDirectory()) scan(file);
		else if (entry.name.endsWith('.html')) {
			if (entry.name === '404.html') continue;
			const html = fs.readFileSync(file, 'utf8');
			for (const match of html.matchAll(/(?:href|src)="(\/[^"]*)"/g)) {
				if (prefixes.some((prefix) => match[1].startsWith(prefix))) invalid.push(`${file}: ${match[1]}`);
			}
			for (const match of html.matchAll(/(?:href|src)="([^"]+)"/g)) {
				if (/^(?:#|mailto:|javascript:)/.test(match[1])) continue;
				const url = new URL(match[1], origin + deployedPath(file));
				if (url.origin !== origin || ['/', '/v1/'].includes(url.pathname)) continue;
				if (!url.pathname.startsWith('/v2/')) {
					invalid.push(`${file}: escapes /v2/: ${match[1]}`);
					continue;
				}
				const target = targetFile(url.pathname);
				if (!fs.existsSync(target)) invalid.push(`${file}: missing ${url.pathname}`);
				else if (url.hash && target.endsWith('.html') && !fs.readFileSync(target, 'utf8').includes(`id="${decodeURIComponent(url.hash.slice(1))}"`)) {
					invalid.push(`${file}: missing ${url.pathname}${url.hash}`);
				}
			}
		}
	}
}

scan(root);
if (invalid.length) {
	console.error(invalid.join('\n'));
	process.exit(1);
}
