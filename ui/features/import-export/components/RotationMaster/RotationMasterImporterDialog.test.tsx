// The dialog's promises: checking changes nothing, replacing changes the rotation and nothing else, and what is not a
// Rotation Master export is reported in the dialog instead of being applied.
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { Class, Spec } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@testing/sim/fake_host';
import { act, fireEvent, render } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { encodeRmExport } from '../../rotation_master';
import { RotationMasterImporterDialog } from './RotationMasterImporterDialog';

const trackPageView = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackPageView }));
vi.mock('@sim/state/batch', () => ({ batch: (body: () => void) => body() }));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));

const rootElem = document.createElement('div');
const sets: APLRotation[] = [];
const player = {
	getClass: () => Class.ClassMage,
	getSpec: () => Spec.SpecMage,
	aplRotation: APLRotation.create({ type: APLRotation_Type.TypeAPL }),
	setAplRotation: (rotation: APLRotation) => void sets.push(rotation),
};
const host = fakeHost({ rootElem, player: player as never, sim: { db: { getSpellIcons: () => [] } } as never });

const exported = (value: object): string => {
	const out = encodeRmExport(value as never);
	if (!out.ok) throw new Error(out.reason);
	return out.text;
};
const GOOD = exported({
	rotation: [
		{ id: 'a', type: 'spell', action: 11129, conditions: { type: 'SPELL_AVAIL', spell: 11129 } },
		{ id: 'b', type: 'spell', action: 133, conditions: { type: 'ZONE', value: 'x' } },
	],
	cooldowns: [],
});

const renderDialog = (onOpenChange = vi.fn()) => {
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={host}>
				<RotationMasterImporterDialog open onOpenChange={onOpenChange} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
	return { onOpenChange };
};
const paste = (text: string) => {
	const textarea = rootElem.querySelector('textarea')!;
	fireEvent.change(textarea, { target: { value: text } });
};
const click = async (testId: string) => {
	await act(async () => {
		fireEvent.click(rootElem.querySelector(`[data-testid="${testId}"]`)!);
	});
};

describe('RotationMasterImporterDialog', () => {
	beforeEach(() => {
		toasts.length = 0;
		sets.length = 0;
		trackPageView.mockClear();
		document.body.appendChild(rootElem);
	});

	it('shows what the export becomes without changing anything', async () => {
		renderDialog();
		paste(GOOD);
		await click('check-button');
		expect(sets).toHaveLength(0);
		// the test i18n answers with the key; the counts go in as its arguments
		expect(rootElem.querySelector('[data-testid="rotation-master-result"]')!.textContent).toContain('import.rotation_master.result');
		// the step that lost a condition is listed with the reason, the exact one is folded away
		expect(rootElem.querySelector('[data-testid="rotation-master-attention-trigger"]')).not.toBe(null);
		expect(rootElem.querySelector('[data-testid="rotation-master-exact"]')).not.toBe(null);
	});

	it('replaces the rotation, reports it, and closes', async () => {
		const { onOpenChange } = renderDialog();
		paste(GOOD);
		await click('import-button');
		expect(sets).toHaveLength(1);
		expect(sets[0].priorityList).toHaveLength(2);
		expect(sets[0].priorityList[1].hide).toBe(false);
		expect(toasts).toEqual([{ variant: 'success', body: 'import.rotation_master.success' }]);
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it('can hide the steps that lost something', async () => {
		renderDialog();
		paste(GOOD);
		await click('import-hidden-button');
		expect(sets[0].priorityList.map(i => i.hide)).toEqual([false, true]);
	});

	it('says why a text is refused and applies nothing', async () => {
		const { onOpenChange } = renderDialog();
		paste('this is not an export');
		await click('import-button');
		expect(sets).toHaveLength(0);
		expect(rootElem.querySelector('[data-testid="rotation-master-error"]')!.textContent).toContain('Rotation Master export');
		expect(onOpenChange).not.toHaveBeenCalled();
	});

	it('forgets a check when the text changes', async () => {
		renderDialog();
		paste(GOOD);
		await click('check-button');
		expect(rootElem.querySelector('[data-testid="rotation-master-result"]')).not.toBe(null);
		paste(GOOD + ' ');
		expect(rootElem.querySelector('[data-testid="rotation-master-result"]')).toBe(null);
	});
});
