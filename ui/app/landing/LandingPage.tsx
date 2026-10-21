import { PlayerClasses } from '@sim/player/classes/index';

import { LANDING_CLASS_ORDER } from './landing_classes';
import { LandingClassCard } from './LandingClassCard';
import { LandingFooter } from './LandingFooter';
import { LandingHeader } from './LandingHeader';
import { LandingHero } from './LandingHero';

export const LandingPage = () => (
	<>
		<div className="fixed -z-1 h-full w-full bg-landing bg-cover bg-center bg-no-repeat opacity-15" />
		<div id="homepage" className="flex min-h-full flex-col">
			<LandingHeader />
			<main className="grow">
				<div className="ui-landing-container flex flex-col pt-12 pb-22">
					<LandingHero />
					<div className="grid auto-rows-fr grid-cols-landing-cards gap-6" id="sim-links">
						{LANDING_CLASS_ORDER.map((classId, index) => (
							<LandingClassCard key={classId} playerClass={PlayerClasses.fromProto(classId)} index={index} />
						))}
					</div>
				</div>
			</main>
			<LandingFooter />
		</div>
	</>
);
