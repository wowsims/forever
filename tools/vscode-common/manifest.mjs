// The package.json VS Code reads from a built extension: the keys it uses, and the make target that writes it.

export const MANIFEST_KEYS = ['name', 'displayName', 'version', 'publisher', 'engines', 'main', 'activationEvents', 'contributes'];

export function generatedManifest(pkg, makeTarget) {
	const manifest = {};
	for (const key of MANIFEST_KEYS) {
		manifest[key] = pkg[key];
	}
	manifest['x-generated'] = `make ${makeTarget}`;
	return manifest;
}

export function manifestText(pkg, makeTarget) {
	return `${JSON.stringify(generatedManifest(pkg, makeTarget), null, '\t')}\n`;
}
