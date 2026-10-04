import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ConnectedIconInput } from './ConnectedIconInput';

vi.mock('../ConnectedIconPicker', () => ({
	ConnectedIconPicker: ({ testId }: { testId?: string }) => <div data-testid="bound-icon-picker" data-forwarded={testId} />,
}));
vi.mock('../ConnectedIconEnumPicker', () => ({
	ConnectedIconEnumPicker: ({ testId }: { testId?: string }) => <div data-testid="bound-icon-enum-picker" data-forwarded={testId} />,
}));

describe('ConnectedIconInput', () => {
	it('renders the icon picker for an icon config', () => {
		render(<ConnectedIconInput modObject={{}} config={{ type: 'icon' } as never} testId="buff" />);
		expect(screen.getByTestId('bound-icon-picker').getAttribute('data-forwarded')).toBe('buff');
		expect(screen.queryByTestId('bound-icon-enum-picker')).toBeNull();
	});

	it('renders the icon enum picker for an iconEnum config', () => {
		render(<ConnectedIconInput modObject={{}} config={{ type: 'iconEnum' } as never} />);
		expect(screen.getByTestId('bound-icon-enum-picker')).toBeTruthy();
		expect(screen.queryByTestId('bound-icon-picker')).toBeNull();
	});
});
