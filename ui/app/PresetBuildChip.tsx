import { Chip } from '@ui-kit/Chip';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';

import type { PresetBuild } from './preset_utils';
import { PresetBuildTooltip } from './PresetBuildTooltip';

export interface PresetBuildChipProps {
	build: PresetBuild;
	active?: boolean;
	/** The id of the `PresetBuildChipTooltip` that describes this build on hover. */
	tooltipId: string;
	onApply: (build: PresetBuild) => void;
}

export const PresetBuildChip = ({ build, active, tooltipId, onApply }: PresetBuildChipProps) => (
	<Chip
		as="button"
		nameAs="span"
		testId="saved-data-set-chip"
		nameTestId="saved-data-set-name"
		label={build.name}
		active={active}
		onSelect={() => onApply(build)}
		nameProps={{ role: 'button' }}
		rootProps={{ type: 'button', ...tooltipAnchorProps(tooltipId, build.name) }}
	/>
);

export interface PresetBuildChipTooltipProps {
	id: string;
	builds: ReadonlyArray<PresetBuild>;
}

/** One tooltip for every chip anchored to `id`; it finds the hovered build by the name the anchor carries. */
export const PresetBuildChipTooltip = ({ id, builds }: PresetBuildChipTooltipProps) => (
	<Tooltip
		id={id}
		render={({ activeAnchor }) => {
			const build = builds.find(candidate => candidate.name === activeAnchor?.getAttribute('data-tooltip-content'));
			return build ? <PresetBuildTooltip build={build} /> : null;
		}}
	/>
);
