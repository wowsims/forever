import type { ReforgeOptimizerModel, ReforgeOptimizerOptions } from '@features/reforge/model/reforge_optimizer';

import { ReforgePanel } from './ReforgePanel';

export interface ReforgeSidebarGroupProps {
	model: ReforgeOptimizerModel;
	options?: ReforgeOptimizerOptions;
}

export const ReforgeSidebarGroup = ({ model, options }: ReforgeSidebarGroupProps) => (
	<div
		role="group"
		className="ui-reforge-sidebar-actions order-20 grid w-fit grid-cols-reforge [--settings-button-width:--spacing(10)]"
		data-testid="suggest-reforges-settings-group">
		<ReforgePanel model={model} options={options} />
	</div>
);
