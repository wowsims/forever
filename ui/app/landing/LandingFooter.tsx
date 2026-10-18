import i18n from '@i18n/config';

export const LandingFooter = () => (
	<footer className="fixed inset-x-0 bottom-0 z-toolbar border-t border-solid border-white/8 bg-background/90 text-[0.8rem] text-white/55">
		<div className="ui-landing-container flex flex-wrap justify-center gap-2 py-2 text-center">
			<span>{i18n.t('landing.footer.copyright', { year: new Date().getFullYear() })}</span>
			<span className="max-sm:hidden" aria-hidden="true">
				·
			</span>
			<span>{i18n.t('landing.footer.disclaimer')}</span>
		</div>
	</footer>
);
