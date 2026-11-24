import clsx from 'clsx';

export interface MeterProps {
	value: number;
	max: number;
	/** `xs` is 1 spacing step tall, `sm` 2. */
	size?: 'xs' | 'sm';
	track?: 'surface' | 'gray';
}

/** A bar filled in the primary colour up to `value` of `max`. It is hidden from assistive tech: every caller writes the count beside it. */
export const Meter = ({ value, max, size = 'sm', track = 'surface' }: MeterProps) => (
	<div aria-hidden className={clsx('overflow-hidden', size === 'xs' ? 'h-1' : 'h-2', track === 'gray' ? 'bg-gray-800' : 'bg-surface')}>
		<div className="h-full bg-primary" style={{ width: `${max ? (100 * Math.min(value, max)) / max : 0}%` }} />
	</div>
);
