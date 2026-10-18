import { getLang, setLang, supportedLanguages } from '@i18n/locale_service';
import { Icon } from '@ui-kit/Icon';
import { Menu, MenuItem } from '@ui-kit/Menu';
import clsx from 'clsx';

export interface LandingLanguageMenuProps {
	triggerClassName?: string;
}

export const LandingLanguageMenu = ({ triggerClassName }: LandingLanguageMenuProps) => {
	const currentLang = getLang();

	const selectLang = (lang: string) => {
		setLang(lang);
		window.location.reload();
	};

	return (
		<Menu
			surface="none"
			trigger={<Icon name="globe" size="2x" />}
			triggerProps={{
				id: 'languageDropdown',
				className: clsx('ui-landing-language-caret flex items-center text-sm whitespace-nowrap text-white', triggerClassName),
				'aria-label': supportedLanguages[currentLang],
			}}
			align="end"
			sideOffset={0}
			keepMounted
			className="ui-landing-language-popup">
			{Object.entries(supportedLanguages).map(([code, name]) => (
				<MenuItem
					key={code}
					layout="block"
					className="clear-both rounded-none no-underline"
					data-active={code === currentLang ? '' : undefined}
					data-lang={code}
					data-testid="dropdown-item"
					onClick={() => selectLang(code)}>
					{name}
				</MenuItem>
			))}
		</Menu>
	);
};
