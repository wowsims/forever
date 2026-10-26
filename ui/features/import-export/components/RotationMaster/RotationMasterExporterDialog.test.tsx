// The export dialog writes the string the importer reads, says what did not carry over, and warns that fidelity may be lost.
import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { Class, Spec } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@testing/sim/fake_host';
import { render } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { decodeRmExport } from '../../rotation_master';
import { RotationMasterExporterDialog } from './RotationMasterExporterDialog';

vi.mock('../../../../tracking/utils', () => ({ trackPageView: vi.fn() }));

const rootElem = document.createElement('div');
const hostFor = (rotation: APLRotation) =>
	fakeHost({
		rootElem,
		player: {
			getClass: () => Class.ClassMage,
			getSpec: () => Spec.SpecMage,
			getRotationType: () => (rotation.type == APLRotation_Type.TypeUnknown ? APLRotation_Type.TypeAPL : rotation.type),
			getResolvedAplRotation: () => rotation,
		} as never,
		sim: { db: { getSpellIcons: () => [] } } as never,
	});

const renderDialog = (rotation: APLRotation) =>
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={hostFor(rotation)}>
				<RotationMasterExporterDialog open onOpenChange={vi.fn()} />
			</SimHostProvider>
		</PortalContainerContext>,
	);

describe('RotationMasterExporterDialog', () => {
	beforeEach(() => {
		document.body.appendChild(rootElem);
	});

	it('shows an export string for the APL, with the fidelity warning and the report', () => {
		renderDialog(
			APLRotation.fromJson({
				type: 'TypeAPL',
				priorityList: [
					{ action: { condition: { spellIsReady: { spellId: { spellId: 11129 } } }, castSpell: { spellId: { spellId: 11129 } } } },
					{
						action: {
							condition: { cmp: { op: 'OpGt', lhs: { remainingTime: {} }, rhs: { const: { val: '10s' } } } },
							castSpell: { spellId: { spellId: 12051 } },
						},
					},
				],
			}),
		);
		const text = rootElem.querySelector<HTMLTextAreaElement>('[data-testid="exporter-textarea"]')!.value;
		const decoded = decodeRmExport(text);
		expect(decoded.ok && (decoded.value as { rotation: unknown[] }).rotation).toHaveLength(2);
		expect(rootElem.textContent).toContain('export.rotation_master.warning_message');
		// the item that lost its fight-length condition is among the steps to check
		expect(rootElem.querySelector('[data-testid="rotation-master-attention-trigger"]')!.textContent).toContain('import.rotation_master.report.attention');
		expect(rootElem.querySelector('[data-testid="rotation-master-error"]')).toBe(null);
	});

	it('says so when there is nothing to export', () => {
		renderDialog(APLRotation.create({ type: APLRotation_Type.TypeAPL }));
		expect(rootElem.querySelector('[data-testid="rotation-master-error"]')!.textContent).toContain('This APL has no actions');
		expect(rootElem.querySelector<HTMLTextAreaElement>('[data-testid="exporter-textarea"]')!.value).toBe('');
	});

	it('says when the exported list was generated from an Auto rotation', () => {
		renderDialog(APLRotation.fromJson({ type: 'TypeAuto', priorityList: [{ action: { castSpell: { spellId: { spellId: 11129 } } } }] }));
		expect(rootElem.textContent).toContain('export.rotation_master.generated');
	});
});
