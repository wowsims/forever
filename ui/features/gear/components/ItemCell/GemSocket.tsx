import type { GemColor } from '@generated/proto/common';
import type { UIGem as Gem } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { getEmptyGemSocketIconUrl } from '@sim/proto/gems';
import { wowheadIconUrl } from '@sim/proto/wowhead';
import { WowheadAnchor, type WowheadAnchorProps } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';

export interface GemSocketProps extends Omit<WowheadAnchorProps, 'href' | 'children'> {
	socketColor: GemColor;
	gem: Gem | null;
}

export const GemSocket = ({ socketColor, gem, className, ...rest }: GemSocketProps) => {
	const href = gem ? ActionId.makeItemUrl(gem.id) : undefined;
	const emptyIconUrl = getEmptyGemSocketIconUrl(socketColor);

	return (
		<WowheadAnchor
			{...rest}
			className={clsx('ui-gem-socket ui-gem-socket-focus-scope relative size-(--gem-width) shrink-0 not-last:mr-px', className)}
			testId="gem-socket-container"
			href={href}>
			{gem && (
				<img
					className="absolute inset-gem z-1 inline-block size-gem-inner cursor-pointer rounded-full bg-cover bg-center bg-no-repeat"
					data-testid="gem-icon"
					src={wowheadIconUrl(gem.icon)}
					alt=""
				/>
			)}
			<img
				className="absolute inset-0 inline-block size-full cursor-pointer bg-cover bg-center bg-no-repeat"
				data-testid="socket-icon"
				src={emptyIconUrl}
				alt=""
			/>
		</WowheadAnchor>
	);
};
