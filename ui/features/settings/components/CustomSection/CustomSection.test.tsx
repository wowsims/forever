import { SimHostProvider } from '@sim/context/SimHostContext';
import type { CustomSection as CustomSectionConfig } from '@sim/spec_config';
import { act, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// The real source selects from a zustand store on `player.sim`; this file is about what the section
// does with the answer, so the source is stubbed and driven directly.
const source = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	return { listeners, notify: () => listeners.forEach(listener => listener()) };
});
vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	const { noopSubscribe } = await import('@testing/sim/fake_host');
	return mockSubscriptions(noopSubscribe, {
		subscribePlayerChange: () => (onChange: () => void) => {
			source.listeners.add(onChange);
			return () => source.listeners.delete(onChange);
		},
	});
});

// Both take a live player and are covered by their own suites.
vi.mock('@features/inputs/components/ConnectedIconInput', () => ({
	ConnectedIconInput: ({ config }: { config: { layout?: string } }) => (
		<div className="icon-picker-root" data-testid="icon-picker-root" data-layout={config.layout} />
	),
}));
vi.mock('@features/inputs/components/ConnectedInput', () => ({
	ConnectedInput: ({ config }: { config: { id: string; layout?: string } }) => (
		<div data-testid="input-picker-stub" data-id={config.id} data-layout={config.layout} />
	),
}));

const { CustomSection } = await import('./CustomSection');

/** Stands in for the player: the one flag the `when` predicates below ask about. */
const player = { shown: true };

const mount = (section: Partial<CustomSectionConfig<any>>) => {
	const { container } = render(
		<SimHostProvider host={{ player } as never}>
			<CustomSection section={{ id: 'totems', title: 'Totems', ...section } as CustomSectionConfig<any>} />
		</SimHostProvider>,
	);
	return container.querySelector('[data-testid="content-block"]') as HTMLElement;
};

describe('CustomSection', () => {
	beforeEach(() => {
		player.shown = true;
		source.listeners.clear();
	});

	it('renders the title into the block header', () => {
		expect(mount({}).querySelector('[data-testid="content-block-title"]')!.textContent).toBe('Totems');
	});

	// The header lays its children out in a row, so a description only lands under the title if the
	// header is also stacked.
	it('renders a description under the title, stacking the header', () => {
		const header = mount({ description: 'The selected curse blocks the matching external debuff.' }).querySelector('[data-testid="content-block-header"]')!;
		expect(header.querySelector('p')!.textContent).toBe('The selected curse blocks the matching external debuff.');
		expect(header.classList.contains('flex-col')).toBe(true);
	});

	it('leaves the header unstacked and adds no paragraph when the section has no description', () => {
		const header = mount({}).querySelector('[data-testid="content-block-header"]')!;
		expect(header.querySelector('p')).toBeNull();
		expect(header.classList.contains('flex-col')).toBe(false);
	});

	it('renders its icons into the block body bare, and forces its inputs inline', () => {
		const block = mount({
			iconInputs: [{ type: 'icon' }] as never,
			inputs: [{ id: 'disable-immolate', type: 'boolean' }] as never,
		});
		const body = block.querySelector('[data-testid="content-block-body"]')!;
		const iconPicker = body.querySelector('[data-testid="icon-picker-root"]')!;
		expect(iconPicker.hasAttribute('data-layout')).toBe(false);
		expect(body.querySelector('[data-testid="input-picker-stub"]')!.getAttribute('data-layout')).toBe('inline');
	});

	it('renders no block at all when `when` is false at mount', () => {
		expect(mount({ when: () => false })).toBeNull();
	});

	it('unmounts and remounts the block as the player changes', () => {
		const block = () => document.querySelector('[data-testid="content-block"]');
		mount({ when: ((subject: typeof player) => subject.shown) as unknown as CustomSectionConfig<any>['when'] });
		expect(block()).toBeTruthy();

		act(() => {
			player.shown = false;
			source.notify();
		});
		expect(block()).toBeNull();

		act(() => {
			player.shown = true;
			source.notify();
		});
		expect(block()).toBeTruthy();
	});
});
