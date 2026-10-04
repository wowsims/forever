import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { PlayerSettingsView } from './PlayerSettings';

const slot = (id: string) => <span key={id} data-testid={id} />;

describe('PlayerSettingsView', () => {
	it('groups the icon pickers, then the race, the spec inputs and the two professions', () => {
		const { container } = render(
			<PlayerSettingsView
				iconPickers={[slot('icon-1'), slot('icon-2')]}
				race={slot('race')}
				inputs={[slot('input-1')]}
				professions={[slot('profession-1'), slot('profession-2')]}
			/>,
		);

		const iconGroup = screen.getByTestId('player-icon-group');
		expect(Array.from(iconGroup.children).map(child => child.getAttribute('data-testid'))).toEqual(['icon-1', 'icon-2']);
		expect(Array.from(container.children).map(child => child.getAttribute('data-testid'))).toEqual(['player-icon-group', 'race', 'input-1', null]);
		expect(Array.from(container.children[3].children).map(child => child.getAttribute('data-testid'))).toEqual(['profession-1', 'profession-2']);
	});

	it('leaves the icon group out when the spec has no icon inputs', () => {
		render(<PlayerSettingsView iconPickers={[]} race={slot('race')} inputs={[]} professions={[]} />);

		expect(screen.queryByTestId('player-icon-group')).toBeNull();
		expect(screen.getByTestId('race')).toBeTruthy();
	});
});
