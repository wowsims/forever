import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { withSimHost } from './models';

const seen: { gear?: unknown; epWeights?: unknown; talentsString?: unknown } = {};

const Probe = () => {
	seen.gear = usePlayerStore('gear');
	seen.epWeights = usePlayerStore('epWeights');
	seen.talentsString = usePlayerStore('talentsString');
	return <span data-testid="probe" />;
};

const renderThrough = (decorator: ReturnType<typeof withSimHost>) => {
	const Decorated = () => decorator(Probe as never, {} as never);
	render(<Decorated />);
};

describe('withSimHost', () => {
	it('seeds the player slice, so usePlayerStore reads the value a playerSlice part hands it', () => {
		const gear = new Gear({});

		renderThrough(withSimHost({ playerSlice: { gear, talentsString: '0123' } }));

		expect(seen.gear).toBe(gear);
		expect(seen.talentsString).toBe('0123');
	});

	it("reads a fresh player's defaults for every field the part leaves out", () => {
		renderThrough(withSimHost());

		expect(seen.gear).toBeInstanceOf(Gear);
		expect(seen.epWeights).toBeInstanceOf(Stats);
		expect(seen.talentsString).toBe('');
	});
});
