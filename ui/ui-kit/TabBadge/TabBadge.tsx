export interface TabBadgeProps {
	label?: string;
	testId?: string;
}

// The parentheses are content, not decoration: they are text nodes either side of the span, and the
// tab strip is compared as text by the parity gates.
export const TabBadge = ({ label, testId = 'tab-badge' }: TabBadgeProps) =>
	label ? (
		<>
			{' ('}
			<span className="text-success" data-testid={testId}>
				{label}
			</span>
			{')'}
		</>
	) : null;
