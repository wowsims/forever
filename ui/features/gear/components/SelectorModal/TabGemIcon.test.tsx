import { GemColor } from '@generated/proto/common';
import type { UIGem as Gem } from '@generated/proto/ui';
import { getEmptyGemSocketIconUrl } from '@sim/proto/gems';
import { wowheadIconUrl } from '@sim/proto/wowhead';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TabGemIcon } from './TabGemIcon';

describe('TabGemIcon', () => {
	const gem = { id: 42, name: 'Gem', icon: 'inv_misc_gem_ruby_01' } as Gem;

	it('shows the gem icon and the socket icon for a filled socket', () => {
		const { container } = render(<TabGemIcon socketColor={GemColor.GemColorRed} gem={gem} />);

		const gemIcon = container.querySelector<HTMLImageElement>('[data-testid="gem-icon"]')!;
		const socketIcon = container.querySelector<HTMLImageElement>('[data-testid="socket-icon"]')!;
		expect(gemIcon.src).toBe(wowheadIconUrl(gem.icon));
		expect(socketIcon.src).toContain(getEmptyGemSocketIconUrl(GemColor.GemColorRed));
	});

	it('renders no gem icon and points the socket image at the empty socket for an empty socket', () => {
		const { container } = render(<TabGemIcon socketColor={GemColor.GemColorRed} gem={null} />);

		const socketIcon = container.querySelector<HTMLImageElement>('[data-testid="socket-icon"]')!;
		const emptyUrl = getEmptyGemSocketIconUrl(GemColor.GemColorRed);
		expect(container.querySelector('[data-testid="gem-icon"]')).toBeNull();
		expect(socketIcon.src).toContain(emptyUrl);
	});

	it('resolves a different empty-socket url for a different socket color', () => {
		const { container: red } = render(<TabGemIcon socketColor={GemColor.GemColorRed} gem={null} />);
		const { container: blue } = render(<TabGemIcon socketColor={GemColor.GemColorBlue} gem={null} />);

		const redSrc = red.querySelector<HTMLImageElement>('[data-testid="socket-icon"]')!.src;
		const blueSrc = blue.querySelector<HTMLImageElement>('[data-testid="socket-icon"]')!.src;
		expect(redSrc).not.toBe(blueSrc);
	});
});
