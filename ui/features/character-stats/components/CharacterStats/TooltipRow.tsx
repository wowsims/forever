import clsx from 'clsx';

export interface TooltipRowProps {
	label: string;
	value: string;
	className?: string;
	cap?: 'met' | 'missed';
}

export const TooltipRow = ({ label, value, className, cap }: TooltipRowProps) => (
	<div data-testid="character-stats-tooltip-row" data-cap={cap} className={clsx('flex justify-between', className)}>
		<span className="mr-2">{label}</span>
		<span>{value}</span>
	</div>
);
