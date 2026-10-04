import { AplList } from '@features/apl/components/AplList';
import { PrePullList } from '@features/apl/components/PrePullList';
import { PriorityList } from '@features/apl/components/PriorityList';
import { APL_ROWS } from '@features/apl/model/apply_move';

/** The first APL sub-tab: the pre-pull list and the priority list share one tree, so a row can be dragged between them. */
export const PriorityPane = () => (
	<AplList.Root group={APL_ROWS}>
		<PrePullList />
		<PriorityList />
	</AplList.Root>
);
