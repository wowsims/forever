import { REPO_RELEASES_URL } from '@sim/constants/other';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { BulkTabActionsView, type BulkTabActionsViewProps } from './BulkTabBody';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const props = (overrides: Partial<BulkTabActionsViewProps> = {}): BulkTabActionsViewProps => ({
	showNativeDownload: false,
	onImportBags: vi.fn(),
	onImportFavorites: vi.fn(),
	onClear: vi.fn(),
	...overrides,
});

describe('BulkTabActionsView', () => {
	it('links the native download only when asked to', () => {
		const { rerender } = render(<BulkTabActionsView {...props()} />);
		expect(screen.queryByRole('link')).toBeNull();

		rerender(<BulkTabActionsView {...props({ showNativeDownload: true })} />);
		expect(screen.getByRole('link', { name: 'bulk_tab.download_native' }).getAttribute('href')).toBe(REPO_RELEASES_URL);
	});

	it('reports each action', () => {
		const handlers = props();
		render(<BulkTabActionsView {...handlers} />);

		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.actions.import_bags' }));
		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.actions.import_favorites' }));
		fireEvent.click(screen.getByRole('button', { name: 'bulk_tab.actions.clear_items' }));

		expect(handlers.onImportBags).toHaveBeenCalledTimes(1);
		expect(handlers.onImportFavorites).toHaveBeenCalledTimes(1);
		expect(handlers.onClear).toHaveBeenCalledTimes(1);
	});
});
