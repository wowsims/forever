import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { useState } from 'react';

import { BulkCompareDialog } from '../BulkCompareDialog';

export interface BulkCompareButtonViewProps {
	onOpen: () => void;
}

export const BulkCompareButtonView = ({ onOpen }: BulkCompareButtonViewProps) => (
	<Button variant="outline-primary" size="sm" onClick={onOpen}>
		<Icon name="scale-balanced" />
		{i18n.t('bulk_tab.compare.button')}
	</Button>
);

export const BulkCompareButton = () => {
	const [open, setOpen] = useState(false);
	return (
		<>
			<BulkCompareButtonView onOpen={() => setOpen(true)} />
			<BulkCompareDialog open={open} onOpenChange={setOpen} />
		</>
	);
};
