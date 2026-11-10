import i18n from '@i18n/config';
import { ADDON_URL, DISCORD_URL, GITHUB_ORG_URL, PATREON_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { useEffect, useState } from 'react';

import { fetchFooterContent, type FooterContent } from './footer_content';
import { LandingBrand } from './LandingBrand';

const SOCIAL_LINKS = [
	{ href: DISCORD_URL, label: 'landing.header.discord' },
	{ href: GITHUB_ORG_URL, label: 'landing.header.github' },
	{ href: ADDON_URL, label: 'landing.footer.addon' },
	{ href: PATREON_URL, label: 'landing.footer.patreon' },
] as const;

const FooterGroup = ({ id, title, split, children }: { id: string; title: string; split?: boolean; children: React.ReactNode }) => (
	<nav aria-labelledby={id} className={split ? 'wowsims-footer-group wowsims-footer-group-split' : 'wowsims-footer-group'}>
		<h2 id={id} className="wowsims-footer-heading">
			{title}
		</h2>
		<ul className="wowsims-footer-links">{children}</ul>
	</nav>
);

/**
 * The footer every WoWSims site shares: its content comes from wowsims.com/versions.json and its styles from
 * wowsims.com/footer.css. Nothing renders until the manifest loads, or when it cannot be read.
 */
export const LandingFooter = () => {
	const [content, setContent] = useState<FooterContent | null>(null);

	useEffect(() => {
		const controller = new AbortController();
		fetchFooterContent(controller.signal).then(fetched => {
			if (fetched && !controller.signal.aborted) setContent(fetched);
		});
		return () => controller.abort();
	}, []);

	if (!content) return null;
	const { legal } = content;

	return (
		<footer className="wowsims-footer">
			<div className="ui-landing-container">
				<div className="wowsims-footer-columns">
					<div className="wowsims-footer-brand">
						<LandingBrand />
						<ul className="wowsims-footer-socials">
							{SOCIAL_LINKS.map(({ href, label }, index) => (
								<li key={href}>
									{index > 0 && (
										<span className="wowsims-footer-dot" aria-hidden="true">
											·
										</span>
									)}
									<Button href={href} target="_blank" variant="unstyled">
										{i18n.t(label)}
									</Button>
								</li>
							))}
						</ul>
					</div>
					<FooterGroup id="landing-footer-sims" title={i18n.t('landing.footer.sims')}>
						{content.versions.map(version => (
							<li key={version.url}>
								{version.status === 'wip' ? (
									<span className="wowsims-footer-wip">{version.title}</span>
								) : (
									<Button href={version.url} variant="unstyled">
										{version.title}
									</Button>
								)}
							</li>
						))}
					</FooterGroup>
					{content.communities.map((group, index) => (
						<FooterGroup key={group.title} id={`landing-footer-community-${index}`} title={group.title} split={group.columns === 2}>
							{group.links.map(link => (
								<li key={link.url}>
									<Button href={link.url} target="_blank" variant="unstyled">
										{link.name}
									</Button>
								</li>
							))}
						</FooterGroup>
					))}
				</div>
				<div className="wowsims-footer-legal">
					<span>
						{i18n.t('landing.footer.copyright', {
							startYear: legal.copyrightStartYear,
							year: new Date().getFullYear(),
							holder: legal.copyrightHolder,
						})}
					</span>
					<span className="wowsims-footer-dot" aria-hidden="true">
						·
					</span>
					<span>{i18n.t('landing.footer.disclaimer')}</span>
				</div>
			</div>
		</footer>
	);
};
