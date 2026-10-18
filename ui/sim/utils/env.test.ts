import { isPreviewOrDevMode } from '@sim/utils/env';
import { afterEach, describe, expect, it, vi } from 'vitest';

describe('isPreviewOrDevMode', () => {
	afterEach(() => {
		vi.unstubAllEnvs();
	});

	it('holds on the dev server whatever the host', () => {
		expect(isPreviewOrDevMode('www.wowsims.com')).toBe(true);
	});

	describe('in a production build', () => {
		const onHost = (hostname: string) => {
			vi.stubEnv('DEV', false);
			return isPreviewOrDevMode(hostname);
		};

		it('holds on the Cloudflare Pages host and its previews', () => {
			expect(onHost('wowsims-forever.pages.dev')).toBe(true);
			expect(onHost('0123456789abcdef.wowsims-forever.pages.dev')).toBe(true);
		});

		it('does not hold anywhere else', () => {
			expect(onHost('www.wowsims.com')).toBe(false);
			expect(onHost('wowsims.com')).toBe(false);
			expect(onHost('other-project.pages.dev')).toBe(false);
			expect(onHost('wowsims-forever.pages.dev.example.com')).toBe(false);
			expect(onHost('notwowsims-forever.pages.dev')).toBe(false);
		});
	});
});
