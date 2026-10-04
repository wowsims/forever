import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

// No pointer to an external healing sim: QE Live covers MoP only.
export const UnlaunchedNotice = () => (
	<div
		className="mt-auto mr-auto mb-auto ml-auto flex max-w-100 flex-col items-center text-center [&_i]:text-danger"
		data-testid="sim-ui-unlaunched-container">
		<Icon name="ban" size="3x" className="mb-2" />
		<h6>{i18n.t('sim.unlaunched.title')}</h6>
		<p>
			{i18n.t('sim.unlaunched.contribute_message')}
			<br />
			{i18n.t('sim.unlaunched.discord_message')}{' '}
			<Button variant="unstyled" href="https://discord.gg/p3DgvmnDCS" target="_blank">
				Discord
			</Button>
			!
		</p>
	</div>
);
