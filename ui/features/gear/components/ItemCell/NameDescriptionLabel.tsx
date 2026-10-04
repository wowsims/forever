import type { ClassValue } from 'clsx';
import clsx from 'clsx';

export interface NameDescriptionLabelProps {
	nameDescription: string;
	className?: ClassValue;
}

export const NameDescriptionLabel = ({ nameDescription, className }: NameDescriptionLabelProps) => (
	<small className={clsx('text-ui text-quality-uncommon', className)} data-testid="heroic-label">
		({nameDescription})
	</small>
);
