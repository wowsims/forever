import { describe, expectTypeOf, it } from 'vitest';

import type { makeMultistatePartyBuffInput, makeTristateDebuffInput, makeTristatePartyBuffInput, makeTristateRaidBuffInput } from './icon_inputs';

type ConfigOf<F extends (...args: any) => any> = Parameters<F>[0];

// `faction` is read off the player's race, and these factories' mod object is the party or the raid,
// which has no player to read it off. Accepting it would only ever drop it silently.
describe('icon input configs', () => {
	it('offers faction only where the mod object is the player', () => {
		expectTypeOf<ConfigOf<typeof makeTristateRaidBuffInput>>().toHaveProperty('faction');

		expectTypeOf<ConfigOf<typeof makeTristatePartyBuffInput>>().not.toHaveProperty('faction');
		expectTypeOf<ConfigOf<typeof makeTristateDebuffInput>>().not.toHaveProperty('faction');
		expectTypeOf<ConfigOf<typeof makeMultistatePartyBuffInput>>().not.toHaveProperty('faction');
	});

	it('takes enableWhen against the mod object the factory reaches through', () => {
		expectTypeOf<ConfigOf<typeof makeTristateDebuffInput>>().toHaveProperty('enableWhen');
		expectTypeOf<NonNullable<ConfigOf<typeof makeTristateDebuffInput>['enableWhen']>>().parameter(0).toHaveProperty('getDebuffs');
	});
});
