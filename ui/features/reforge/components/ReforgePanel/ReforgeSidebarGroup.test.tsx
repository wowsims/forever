import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('./ReforgePanel', () => ({
	ReforgePanel: () => (
		<>
			<button type="button">Suggest Reforges</button>
			<button type="button">Settings</button>
		</>
	),
}));

const { ReforgeSidebarGroup } = await import('./ReforgeSidebarGroup');

describe('ReforgeSidebarGroup', () => {
	it('groups the run button and the settings trigger', () => {
		render(<ReforgeSidebarGroup model={{} as any} />);

		const group = screen.getByRole('group');
		expect(group).toBe(screen.getByTestId('suggest-reforges-settings-group'));
		expect(
			within(group)
				.getAllByRole('button')
				.map(button => button.textContent),
		).toEqual(['Suggest Reforges', 'Settings']);
	});
});
