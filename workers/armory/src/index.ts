// Looks up a character on the Blizzard Profile API for the sim's Armory importer, holding the app's
// Battle.net credentials so the browser never sees them. See README.md.

import { fetchProfile, UpstreamError } from './blizzard';
import type { ProfileEquipment, ProfileProfessions, ProfileSpecializations, ProfileSummary } from './character';
import { BadRequestError, characterProfilePath, parseCharacterPath, trimCharacter } from './character';

// Only the local dev server for now: the importer is dev-only until the Forever namespace is known.
const ALLOWED_ORIGIN = /^http:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/;

const corsHeaders = (request: Request): Record<string, string> => {
	const origin = request.headers.get('Origin');
	if (!origin || !ALLOWED_ORIGIN.test(origin)) return {};
	return { 'Access-Control-Allow-Origin': origin, 'Access-Control-Allow-Methods': 'GET, OPTIONS', Vary: 'Origin' };
};

const json = (request: Request, status: number, body: unknown): Response => Response.json(body, { status, headers: corsHeaders(request) });

export default {
	async fetch(request, env): Promise<Response> {
		if (request.method == 'OPTIONS') return new Response(null, { status: 204, headers: corsHeaders(request) });
		if (request.method != 'GET') return json(request, 405, { error: 'Method not allowed' });

		try {
			const character = parseCharacterPath(new URL(request.url).pathname);
			const base = characterProfilePath(character, env.NAME_PART_SEPARATOR);
			const [summary, equipment, specializations, professions] = await Promise.all([
				fetchProfile<ProfileSummary>(env, character.region, base),
				fetchProfile<ProfileEquipment>(env, character.region, `${base}/equipment`),
				fetchProfile<ProfileSpecializations>(env, character.region, `${base}/specializations`),
				fetchProfile<ProfileProfessions>(env, character.region, `${base}/professions`),
			]);
			if (!summary) return json(request, 404, { error: 'Character not found' });

			return json(request, 200, trimCharacter(summary, equipment ?? {}, specializations ?? {}, professions));
		} catch (error) {
			if (error instanceof BadRequestError) return json(request, 400, { error: error.message });
			if (error instanceof UpstreamError) {
				console.error(JSON.stringify({ message: error.message, status: error.status }));
				return json(request, 502, { error: error.message });
			}
			throw error;
		}
	},
} satisfies ExportedHandler<Env>;
