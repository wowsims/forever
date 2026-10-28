import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@testing/sim/fake_host';
import { act, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DistanceFromTarget, RetributionAuraSpellPower, ThornsSpellPower } from './other_inputs';

const source = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	return {
		subscribe: (onChange: () => void) => {
			listeners.add(onChange);
			return () => listeners.delete(onChange);
		},
		notify: () => listeners.forEach(listener => listener()),
		clear: () => listeners.clear(),
	};
});

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions(source.subscribe));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

let distance = 0;

const mount = () => {
	const player = {
		getDistanceFromTarget: () => distance,
		setDistanceFromTarget: (value: number) => {
			distance = value;
			source.notify();
		},
	} as any;
	render(
		<SimHostProvider host={fakeHost({ player })}>
			<ConnectedNumberField modObject={player} config={DistanceFromTarget} />
		</SimHostProvider>,
	);
};

const input = () => screen.getByRole('textbox') as HTMLInputElement;

// A real `change` Event, not `fireEvent`, because the picker reads the native value the user
// typed — and going around `fireEvent` goes around act too. The commit notifies the source, which
// re-renders the picker, so the dispatch is wrapped rather than the assertions that follow it.
const commit = (text: string) => {
	input().value = text;
	act(() => void input().dispatchEvent(new Event('change', { bubbles: true })));
};

beforeEach(() => {
	source.clear();
	distance = 20;
});

describe('DistanceFromTarget', () => {
	it('keeps the decimals the user typed', () => {
		mount();
		expect(input().value).toBe('20.00');

		commit('5.5');
		expect(distance).toBe(5.5);
		expect(input().value).toBe('5.50');
	});

	// A negative distance is accepted by the sim and then persisted into saved settings and share
	// links, so the picker is what has to reject it.
	it('drops the sign on a negative distance', () => {
		mount();

		commit('-5');
		expect(distance).toBe(5);
	});
});

// The number only means something on a tank that has the aura ticked, and it lives on the party's
// buffs so the sim reads it next to the aura itself.
describe('RetributionAuraSpellPower', () => {
	const fakePlayer = (isTankSpec: boolean, retributionAura: boolean) => {
		let buffs = { retributionAura, retributionAuraSpellPower: 0 } as any;
		return {
			getPlayerSpec: () => ({ isTankSpec }),
			getParty: () => ({
				getBuffs: () => buffs,
				setBuffs: (next: any) => {
					buffs = next;
				},
			}),
		} as any;
	};

	it('shows only for a tank with the aura selected', () => {
		expect(RetributionAuraSpellPower.showWhen(fakePlayer(true, true))).toBe(true);
		expect(RetributionAuraSpellPower.showWhen(fakePlayer(true, false))).toBe(false);
		expect(RetributionAuraSpellPower.showWhen(fakePlayer(false, true))).toBe(false);
	});

	it('stores the value on the party buffs', () => {
		const player = fakePlayer(true, true);
		RetributionAuraSpellPower.setValue(player, 450);
		expect(player.getParty().getBuffs().retributionAuraSpellPower).toBe(450);
		expect(RetributionAuraSpellPower.getValue(player)).toBe(450);
	});
});

// The same rule for Thorns, whose number lives on the player's own buffs next to the Thorns toggle.
describe('ThornsSpellPower', () => {
	const fakePlayer = (isTankSpec: boolean, thorns: boolean) => {
		let buffs = { thorns, thornsSpellPower: 0 } as any;
		return {
			getPlayerSpec: () => ({ isTankSpec }),
			getBuffs: () => buffs,
			setBuffs: (next: any) => {
				buffs = next;
			},
		} as any;
	};

	it('shows only for a tank with Thorns selected', () => {
		expect(ThornsSpellPower.showWhen(fakePlayer(true, true))).toBe(true);
		expect(ThornsSpellPower.showWhen(fakePlayer(true, false))).toBe(false);
		expect(ThornsSpellPower.showWhen(fakePlayer(false, true))).toBe(false);
	});

	it('stores the value on the individual buffs', () => {
		const player = fakePlayer(true, true);
		ThornsSpellPower.setValue(player, 450);
		expect(player.getBuffs().thornsSpellPower).toBe(450);
		expect(ThornsSpellPower.getValue(player)).toBe(450);
	});
});
