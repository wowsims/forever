import i18n from '@i18n/config';

import { buildCategories } from './preset_build_state';
import type { PresetBuild } from './preset_utils';

export const PresetBuildTooltip = ({ build }: { build: PresetBuild }) => (
	<>
		<p className="mb-1">{i18n.t('common.preset.description')}</p>
		<ul className="mb-0 list-disc pl-4">
			{buildCategories(build).map(category => (
				<li key={category}>{category}</li>
			))}
		</ul>
	</>
);
