import i18n from '@i18n/config';
import { LocaleHtml } from '@ui-kit/Tooltip';
import { TooltipButton } from '@ui-kit/TooltipButton';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ElementType, ReactNode } from 'react';

export interface ContentBlockHeaderProps {
	title: string;
	className?: ClassValue;
	titleTag?: string;
	tooltip?: string;
}

export interface ContentBlockConfigProps {
	bodyClassName?: ClassValue;
	/** For a block whose body would be empty: the body element is not rendered at all. */
	withoutBody?: boolean;
	header?: ContentBlockHeaderProps;
}

export interface ContentBlockProps {
	className?: ClassValue;
	config: ContentBlockConfigProps;
	children?: ReactNode;
	headerChildren?: ReactNode;
	flush?: boolean;
	testId?: string;
	rootDataAttributes?: Record<string, string>;
}

export const ContentBlock = ({ className, config, children, headerChildren, flush, testId = 'content-block', rootDataAttributes }: ContentBlockProps) => {
	const header = config.header;
	const hasHeader = !!header && Object.keys(header).length > 0;
	const TitleTag = (header?.titleTag || 'h6') as ElementType;

	return (
		<div className={clsx('ui-content-block', className, flush && 'mb-0', 'flex flex-col')} data-testid={testId} {...rootDataAttributes}>
			{hasHeader && header && (
				<div className={clsx('ui-content-block-header flex items-baseline gap-2', header.className)} data-testid="content-block-header">
					<TitleTag className="mb-0 flex items-center font-bold" data-testid="content-block-title">
						{header.title}
						{header.tooltip && <TooltipButton tooltip={<LocaleHtml html={header.tooltip} />} label={i18n.t('common.help')} className="ml-2" />}
					</TitleTag>
					{headerChildren}
				</div>
			)}
			{!config.withoutBody && (
				<div className={clsx('ui-content-block-body flex-col', config.bodyClassName)} data-testid="content-block-body">
					{children}
				</div>
			)}
		</div>
	);
};
