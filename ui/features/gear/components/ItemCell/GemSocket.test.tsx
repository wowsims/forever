import { GemColor } from '@generated/proto/common';
import type { UIGem as Gem } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { getEmptyGemSocketIconUrl } from '@sim/proto/gems';
import { wowheadIconUrl } from '@sim/proto/wowhead';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { GemSocket } from './GemSocket';

describe('GemSocket', () => {
	const gem = { id: 42, name: 'Gem', icon: 'inv_misc_gem_ruby_01' } as Gem;

	it('shows the gem icon and links to the gem for a filled socket', () => {
		const { container } = render(<GemSocket socketColor={GemColor.GemColorRed} gem={gem} />);

		const gemIcon = container.querySelector<HTMLImageElement>('[data-testid="gem-icon"]')!;
		const socketIcon = container.querySelector<HTMLImageElement>('[data-testid="socket-icon"]')!;
		expect(gemIcon.src).toBe(wowheadIconUrl(gem.icon));
		expect(container.querySelector('a')!.href).toBe(ActionId.makeItemUrl(gem.id));
		expect(socketIcon.src).toContain(getEmptyGemSocketIconUrl(GemColor.GemColorRed));
	});

	it('renders no gem icon and no href for an empty socket', () => {
		const { container } = render(<GemSocket socketColor={GemColor.GemColorRed} gem={null} />);

		expect(container.querySelector('[data-testid="gem-icon"]')).toBeNull();
		expect(container.querySelector('a')!.hasAttribute('href')).toBe(false);
	});
});
