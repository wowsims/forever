// URL-hash link parsing for individual sim settings, extracted from
// IndividualLinkImporter so domain code (reforge_cache) can use it without
// depending on the view layer.
import { IndividualSimSettings } from '@generated/proto/ui';

import { SIM_CATEGORY_OF_KEY, SimSettingCategories } from '../constants/sim_settings';
import { getEnumValues } from '../utils/collections';

export const LINK_CATEGORY_PARAM = 'i';
export const LINK_DEFAULT_CATEGORIES = getEnumValues(SimSettingCategories).filter(c => c != SimSettingCategories.UISettings) as Array<SimSettingCategories>;

export interface UrlParseData {
	settings: IndividualSimSettings;
	categories: Array<SimSettingCategories>;
}

// Structural subset of `Location` / `URL` / `Env['location']`.
export interface UrlLocation {
	readonly hash: string;
	readonly search: string;
}

// Bytes per `String.fromCharCode` call, far below the engines' argument limits.
const CHAR_CODE_CHUNK = 0x8000;

// A link's hash is the settings' binary form as zlib-wrapped deflate (the 'deflate' format, not 'deflate-raw'),
// then standard base64. `wowsimcli decodelink` reads the same.
export async function encodeLinkHash(bytes: Uint8Array): Promise<string> {
	// `slice()` types the bytes as ArrayBuffer-backed, which a BlobPart needs.
	const deflatedStream = new Blob([bytes.slice()]).stream().pipeThrough(new CompressionStream('deflate'));
	const deflated = new Uint8Array(await new Response(deflatedStream).arrayBuffer());
	let binary = '';
	for (let i = 0; i < deflated.length; i += CHAR_CODE_CHUNK) {
		binary += String.fromCharCode(...deflated.subarray(i, i + CHAR_CODE_CHUNK));
	}
	return btoa(binary);
}

export async function decodeLinkHash(hash: string): Promise<Uint8Array> {
	const binary = atob(hash);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < bytes.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	const inflated = new Blob([bytes]).stream().pipeThrough(new DecompressionStream('deflate'));
	return new Uint8Array(await new Response(inflated).arrayBuffer());
}

// `loc` is any location-shaped source: `sim.env.location`, `window.location`
// or a parsed `URL`.
export async function tryParseUrlLocation(loc: UrlLocation): Promise<UrlParseData | null> {
	let hash = loc.hash;
	if (hash.length <= 1) {
		return null;
	}

	// Remove leading '#'
	hash = hash.substring(1);
	const settings = IndividualSimSettings.fromBinary(await decodeLinkHash(hash));

	let exportCategories = LINK_DEFAULT_CATEGORIES;
	const urlParams = new URLSearchParams(loc.search);
	if (urlParams.has(LINK_CATEGORY_PARAM)) {
		const categoryChars = urlParams.get(LINK_CATEGORY_PARAM)!.split('');
		exportCategories = categoryChars.map(char => SIM_CATEGORY_OF_KEY.get(char)).filter((c): c is SimSettingCategories => c !== undefined);
	}

	return {
		settings: settings,
		categories: exportCategories,
	};
}
