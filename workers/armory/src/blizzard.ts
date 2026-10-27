// Battle.net OAuth and Profile API calls. The token is a client-credentials token, which serves every region.

const TOKEN_URL = 'https://oauth.battle.net/token';
// Renew this long before the token's own expiry, so a request never goes out with one about to lapse.
const TOKEN_EXPIRY_MARGIN_MS = 5 * 60 * 1000;

export class UpstreamError extends Error {
	constructor(
		message: string,
		readonly status: number,
	) {
		super(message);
	}
}

// Shared by every request this isolate serves; it holds no request data, only the app's own token.
let cachedToken: { value: string; expiresAt: number } | null = null;

const getAccessToken = async (env: Env): Promise<string> => {
	if (cachedToken && cachedToken.expiresAt > Date.now()) return cachedToken.value;

	const response = await fetch(TOKEN_URL, {
		method: 'POST',
		headers: {
			Authorization: `Basic ${btoa(`${env.BLIZZARD_CLIENT_ID}:${env.BLIZZARD_CLIENT_SECRET}`)}`,
			'Content-Type': 'application/x-www-form-urlencoded',
		},
		body: 'grant_type=client_credentials',
	});
	if (!response.ok) throw new UpstreamError(`Battle.net token request failed (${response.status})`, response.status);

	const body = await response.json<{ access_token: string; expires_in: number }>();
	cachedToken = { value: body.access_token, expiresAt: Date.now() + body.expires_in * 1000 - TOKEN_EXPIRY_MARGIN_MS };
	return body.access_token;
};

// A 404 resolves to null, for the endpoints a character may lack (professions); any other failure throws.
export const fetchProfile = async <T>(env: Env, region: string, path: string): Promise<T | null> => {
	const token = await getAccessToken(env);
	const url = new URL(path, `https://${region}.api.blizzard.com`);
	url.searchParams.set('namespace', `${env.PROFILE_NAMESPACE_PREFIX}-${region}`);
	url.searchParams.set('locale', 'en_US');

	const response = await fetch(url, { headers: { Authorization: `Bearer ${token}` } });
	if (response.status == 404) return null;
	if (!response.ok) throw new UpstreamError(`Profile API ${url.pathname} failed (${response.status})`, response.status);
	return response.json<T>();
};
