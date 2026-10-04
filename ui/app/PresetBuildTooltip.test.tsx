import { SavedTalents } from '@generated/proto/ui';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { buildCategories } from './preset_build_state';
import type { PresetBuild } from './preset_utils';
import { PresetBuildTooltip } from './PresetBuildTooltip';

describe('PresetBuildTooltip', () => {
	it('lists one item per category the build sets', () => {
		const build = { name: 'Phase 1', talents: { name: 'Fury', data: SavedTalents.create() }, epWeights: {} } as unknown as PresetBuild;

		render(<PresetBuildTooltip build={build} />);

		expect(buildCategories(build)).toHaveLength(2);
		expect(screen.getAllByRole('listitem').map(item => item.textContent)).toEqual(buildCategories(build));
	});
});
