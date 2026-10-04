import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { RotationIconGroup } from './RotationIconGroup';

vi.mock('@sim/context/SimHostContext', () => ({ usePlayer: () => ({}) }));
vi.mock('@features/inputs/components/ConnectedIconPicker', () => ({ ConnectedIconPicker: () => <div data-testid="icon-picker-root" /> }));
vi.mock('@features/inputs/components/ConnectedIconEnumPicker', () => ({ ConnectedIconEnumPicker: () => <div data-testid="icon-enum-picker-root" /> }));

const icon = (type: 'icon' | 'iconEnum') => ({ type }) as never;

describe('RotationIconGroup', () => {
	const group = () => document.querySelector('[data-testid="rotation-icon-group"]') as HTMLElement;

	it('picks the picker each input type names, in order', () => {
		render(<RotationIconGroup inputs={[icon('icon'), icon('iconEnum'), icon('icon')]} />);

		expect([...group().children].map(child => child.getAttribute('data-testid'))).toEqual([
			'icon-picker-root',
			'icon-enum-picker-root',
			'icon-picker-root',
		]);
	});
});
