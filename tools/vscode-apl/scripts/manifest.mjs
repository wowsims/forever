import * as common from '../../vscode-common/manifest.mjs';

export const MANIFEST_KEYS = common.MANIFEST_KEYS;
export const generatedManifest = pkg => common.generatedManifest(pkg, 'vscode-apl');
export const manifestText = pkg => common.manifestText(pkg, 'vscode-apl');
