/** Attributes wowhead's tooltip script reads. `icon: false` for an element with no href: there is no icon to suppress. `acts: false` for a link that only navigates, which keeps wowhead's touch tooltip on a tap. */
export const wowheadAnchorProps = ({ icon = true, acts = true }: { icon?: boolean; acts?: boolean } = {}) => ({
	...(icon ? { 'data-whtticon': 'false' } : {}),
	...(acts ? { 'data-disable-wowhead-touch-tooltip': 'true' } : {}),
});
