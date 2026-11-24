import { APLRotation } from '@generated/proto/apl';
import { RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { UnitStat } from '@sim/proto/stats';
import { renderWithHost } from '@testing/react';
import { filled } from '@testing/sim/hooks';
import { makeIndividualSim, serveRealDatabase } from '@testing/sim/individual_sim';
import { act, screen, waitFor } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { ConsumesPicker } from './ConsumesPicker';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const ELIXIR_OF_FEROCITY = 250350;
const ELIXIR_OF_GREATER_AGILITY = 9187;
const ELIXIR_OF_THE_GRIZZLY = 250351;
const JUJU_POWER = 12451;
const BOGLING_ROOT = 5206;
const ELIXIR_OF_MINOR_FORCE = 247755;
const SCROLL_OF_STRENGTH_IV = 10310;
const SCROLL_OF_INTELLECT_IV = 10308;
const STRENGTH_AND_AGILITY = {
	autoRotation: () => APLRotation.create(),
	epStats: [Stat.StatStrength, Stat.StatAgility],
	displayStats: [UnitStat.fromStat(Stat.StatStrength), UnitStat.fromStat(Stat.StatAgility)],
	presets: { epWeights: [] },
};

// The test setup loads no locale, so the tooltips would read as their keys without the names.
beforeAll(() => {
	serveRealDatabase();
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const renderPicker = async (consumes: ConsumesSpec) => {
	const { player, host } = await makeIndividualSim({ config: STRENGTH_AND_AGILITY });
	player.setConsumes(consumes);
	vi.spyOn(ActionId.prototype, 'fill').mockImplementation(function (this: ActionId) {
		return Promise.resolve(filled(this, String(this.itemId), 'icon.jpg'));
	});
	const { container } = renderWithHost(<ConsumesPicker />, host as unknown as Record<string, unknown>);
	const elixirs = () => screen.getByRole('group', { name: 'Flask & Elixirs' });
	const scrolls = () => container.querySelector('[data-testid="consumes-scrolls"]')!;
	const swatch = (scope: () => ParentNode, itemId: number) =>
		waitFor(() => {
			const anchor = scope().querySelector(`a[href="${ActionId.makeItemUrl(itemId)}"]`);
			expect(anchor).not.toBeNull();
			return anchor!;
		});
	// The swatch's badge tooltip and whether the swatch is dimmed, or undefined with no badge.
	const badgeOn = (anchor: Element) => {
		const box = anchor.parentElement!;
		const tooltip = box.querySelector(':scope > [data-testid="icon-swatch-overwritten"]')?.getAttribute('data-tooltip-content');
		return tooltip ? { tooltip, dimmed: box.hasAttribute('data-overwritten') } : undefined;
	};
	const badge = async (scope: () => ParentNode, itemId: number) => badgeOn(await swatch(scope, itemId));
	// The badge of the miscellaneous row's Offensive icon, which shows its first picked consumable.
	const offensiveBadge = async () => badgeOn(await screen.findByTestId('consumes-misc-offensive'));
	return { player, elixirs, scrolls, swatch, badge, offensiveBadge };
};

describe('ConsumesPicker', () => {
	it('offers the crit elixirs in a dropdown of their own, so Elixir of the Grizzly goes beside Juju Power from the miscellaneous row', async () => {
		const { elixirs, badge, offensiveBadge } = await renderPicker(ConsumesSpec.create({ elixirIds: [ELIXIR_OF_THE_GRIZZLY], extraIds: [JUJU_POWER] }));
		expect(await badge(elixirs, ELIXIR_OF_THE_GRIZZLY)).toEqual({ tooltip: 'Strength overwritten by Juju Power', dimmed: false });
		expect(await offensiveBadge()).toBeUndefined();
	});

	it('dims Bogling Root in the miscellaneous row beside Elixir of Minor Force, which grants as much', async () => {
		const { offensiveBadge } = await renderPicker(ConsumesSpec.create({ elixirIds: [ELIXIR_OF_MINOR_FORCE], extraIds: [BOGLING_ROOT] }));
		expect(await offensiveBadge()).toEqual({
			tooltip: 'Bonus Weapon Damage overwritten by Elixir of Minor Force',
			dimmed: true,
		});
	});

	it('dims an elixir pick every stat of which the other picks match, naming them', async () => {
		const { elixirs, badge } = await renderPicker(
			ConsumesSpec.create({ elixirIds: [ELIXIR_OF_FEROCITY, ELIXIR_OF_GREATER_AGILITY, ELIXIR_OF_THE_GRIZZLY] }),
		);
		expect(await badge(elixirs, ELIXIR_OF_FEROCITY)).toEqual({
			tooltip: 'Strength overwritten by Elixir of the Grizzly\nAgility overwritten by Elixir of Greater Agility',
			dimmed: true,
		});
	});

	it('dims the picked scroll beside a stronger consumable of its stat that counts with the elixirs', async () => {
		const { badge, scrolls, offensiveBadge } = await renderPicker(ConsumesSpec.create({ extraIds: [SCROLL_OF_STRENGTH_IV, JUJU_POWER] }));
		expect(await offensiveBadge()).toBeUndefined();
		expect(await badge(scrolls, SCROLL_OF_STRENGTH_IV)).toEqual({ tooltip: 'Strength overwritten by Juju Power', dimmed: true });
	});

	it('dims the picked scroll once a raid buff of its stat that grants as much is turned on', async () => {
		const { player, badge, scrolls } = await renderPicker(ConsumesSpec.create({ extraIds: [SCROLL_OF_INTELLECT_IV] }));
		act(() => player.getRaid()!.setBuffs(RaidBuffs.create()));
		expect(await badge(scrolls, SCROLL_OF_INTELLECT_IV)).toBeUndefined();

		act(() => player.getRaid()!.setBuffs(RaidBuffs.create({ arcaneIntellect: true })));
		await waitFor(async () =>
			expect(await badge(scrolls, SCROLL_OF_INTELLECT_IV)).toEqual({
				tooltip: 'Intellect overwritten by Arcane Intellect',
				dimmed: true,
			}),
		);
	});
});
