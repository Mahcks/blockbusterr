import fs from 'node:fs';
import path from 'node:path';

// Legacy Markdown links were absolute before the docs supported versioned paths.
function archive(directory) {
	for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
		const file = path.join(directory, entry.name);
		if (entry.isDirectory()) archive(file);
		else if (entry.name.endsWith('.html')) {
			const html = fs.readFileSync(file, 'utf8').replaceAll(
				'ghcr.io/mahcks/blockbusterr:latest', 'ghcr.io/mahcks/blockbusterr:v1.5.2',
			);
			fs.writeFileSync(file, html.replace(
				/href="\/(getting-started|concepts|integrations|api|examples)\//g,
				'href="/v1/$1/',
			));
		}
	}
}

archive(process.argv[2]);
