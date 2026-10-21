import i18n from '@i18n/config';
import { ADDON_URL, DISCORD_URL, GITHUB_ORG_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

export const LandingActions = () => (
	<div className="flex flex-wrap gap-4" data-testid="landing-actions">
		<Button
			href={DISCORD_URL}
			target="_blank"
			variant="unstyled"
			title={i18n.t('landing.actions.discord')}
			className="ui-landing-action ui-landing-action-discord">
			<Icon name="discord" style="brands" />
			{i18n.t('landing.actions.discord')}
		</Button>
		<Button href={GITHUB_ORG_URL} target="_blank" variant="unstyled" title={i18n.t('landing.actions.github')} className="ui-landing-action">
			<Icon name="github" style="brands" />
			{i18n.t('landing.actions.github')}
		</Button>
		<Button
			href={ADDON_URL}
			target="_blank"
			variant="unstyled"
			title={i18n.t('landing.header.addon')}
			className="ui-landing-action ui-landing-action-addon">
			<Icon name="wow" style="brands" />
			{i18n.t('landing.actions.addon')}
		</Button>
	</div>
);
