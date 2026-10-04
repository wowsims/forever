import i18n from '@i18n/config';
import type { Social } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

export interface SocialLinkProps {
	social: Social;
	testId?: string;
}

export const SocialLink = ({ social, testId }: SocialLinkProps) => {
	const id = useId();
	const tooltip = i18n.t(social.tooltip);
	return (
		<>
			<Button
				variant="unstyled"
				href={social.href}
				target="_blank"
				className={social.className}
				aria-label={tooltip}
				data-testid={testId}
				{...tooltipAnchorProps(id)}>
				<Icon name={social.icon} style="brands" size="lg" />
				{'label' in social && social.label}
			</Button>
			<Tooltip id={id} content={tooltip} />
		</>
	);
};
