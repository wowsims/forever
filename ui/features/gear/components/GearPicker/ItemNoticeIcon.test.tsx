import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@testing/sim/fake_host';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { ITEM_INFO_NOTICES, ITEM_NOTICES } from '../../item_notices';
import { ItemNoticeIcon, ItemNoticeIconView } from './ItemNoticeIcon';

const SPEC_SPECIFIC_ITEM = 90101;
const NOTICED_ITEM = 90102;

const mount = (itemId: number, additionalNotice?: ReactNode, spec: Spec = Spec.SpecUnknown, enchantId?: number) => {
	const host = fakeHost({ player: { getSpec: () => spec } });
	return render(
		<SimHostProvider host={host}>
			<ItemNoticeIcon itemId={itemId} enchantId={enchantId} additionalNotice={additionalNotice} />
		</SimHostProvider>,
	);
};

const open = () => fireEvent.mouseEnter(screen.getByRole('button'));

describe('ItemNoticeIcon', () => {
	beforeEach(() => ITEM_NOTICES.set(NOTICED_ITEM, { [Spec.SpecUnknown]: <p>Item notice</p> }));

	afterEach(() => {
		ITEM_NOTICES.delete(SPEC_SPECIFIC_ITEM);
		ITEM_NOTICES.delete(NOTICED_ITEM);
	});

	it('renders nothing for an item with no notice', () => {
		const { container } = mount(1);
		expect(container.firstChild).toBeNull();
	});

	it('shows the item notice behind the warning icon', async () => {
		mount(NOTICED_ITEM);

		expect(screen.getByRole('button').dataset.level).toBe('warning');
		open();
		expect(await screen.findByText('Item notice')).toBeTruthy();
	});

	it('shows an additional notice for an item that has none of its own', async () => {
		mount(1, <p>Please select a random suffix</p>);

		open();
		expect(await screen.findByText('Please select a random suffix')).toBeTruthy();
	});

	it('shows the item notice and the additional notice together, in that order', async () => {
		mount(NOTICED_ITEM, <p>Please select a random suffix</p>);

		open();
		const suffix = await screen.findByText('Please select a random suffix');
		const paragraphs = [...suffix.parentElement!.children].map(child => child.textContent);
		expect(paragraphs).toEqual(['Item notice', 'Please select a random suffix']);
	});

	it('shows an info notice behind the info icon, and the warning icon once a warning joins it', async () => {
		ITEM_INFO_NOTICES.set(SPEC_SPECIFIC_ITEM, <p>Only in Forest and Grassland areas</p>);
		const { rerender, unmount } = mount(SPEC_SPECIFIC_ITEM);

		expect(screen.getByRole('button').dataset.level).toBe('info');
		open();
		expect(await screen.findByText('Only in Forest and Grassland areas')).toBeTruthy();

		rerender(
			<SimHostProvider host={fakeHost({ player: { getSpec: () => Spec.SpecUnknown } })}>
				<ItemNoticeIcon itemId={SPEC_SPECIFIC_ITEM} additionalNotice={<p>Please select a random suffix</p>} />
			</SimHostProvider>,
		);
		expect(screen.getByRole('button').dataset.level).toBe('warning');
		unmount();
		ITEM_INFO_NOTICES.delete(SPEC_SPECIFIC_ITEM);
	});

	it('shows the notice of an enchant whose effect is missing, on an item that has none of its own', async () => {
		mount(1, undefined, Spec.SpecUnknown, 1894);

		expect(screen.getByRole('button').dataset.level).toBe('warning');
		open();
		expect(await screen.findByText('gear_tab.item_notices.missing_effect.enchant_listed')).toBeTruthy();
		expect(await screen.findByText(/often chill the target/)).toBeTruthy();
	});

	it('renders nothing for an enchant whose effect is modelled', () => {
		const { container } = mount(1, undefined, Spec.SpecUnknown, 34);
		expect(container.firstChild).toBeNull();
	});

	it('prefers the player’s spec notice over the generic one', async () => {
		ITEM_NOTICES.set(SPEC_SPECIFIC_ITEM, {
			[Spec.SpecUnknown]: <p>Generic notice</p>,
			[Spec.SpecDpsWarrior]: <p>Dps notice</p>,
		});
		mount(SPEC_SPECIFIC_ITEM, undefined, Spec.SpecDpsWarrior);

		open();
		expect(await screen.findByText('Dps notice')).toBeTruthy();
		expect(screen.queryByText('Generic notice')).toBeNull();
	});
});

describe('ItemNoticeIconView', () => {
	it.each([
		['warning', 'common.list_picker.warnings'],
		['info', 'common.list_picker.additional_information'],
	] as const)('names the icon of a %s notice for its level', (level, name) => {
		render(<ItemNoticeIconView level={level}>Not implemented</ItemNoticeIconView>);
		expect(screen.getByRole('button', { name: i18n.t(name) })).toBeTruthy();
	});

	it('shows its notices in a tooltip on the icon', async () => {
		render(
			<ItemNoticeIconView level="warning">
				<p>First</p>
				<p>Second</p>
			</ItemNoticeIconView>,
		);
		open();
		const first = await screen.findByText('First');
		expect([...first.parentElement!.children].map(child => child.textContent)).toEqual(['First', 'Second']);
	});
});
