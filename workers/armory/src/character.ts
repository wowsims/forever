// The request path and the trimmed character the Worker returns. Pure, so the tests run without the
// Workers runtime.

export const REGIONS = ['us', 'eu', 'kr', 'tw'] as const;
export type Region = (typeof REGIONS)[number];

export interface CharacterRequest {
	region: Region;
	realmSlug: string;
	// Name parts, lowercased: two for a Forever name (Erlyn-Farthing), one for a Classic name.
	nameParts: string[];
}

// What the sim's Armory importer reads. Blizzard's own shapes stay inside this Worker.
export interface ArmoryCharacter {
	name: string;
	realm: string;
	level: number;
	classId: number;
	raceName: string;
	items: Array<{ slot: string; id: number; enchant?: number }>;
	// `talentId` is Blizzard's talent id, which may or may not be the game files' TraitDefinition id.
	talents: Array<{ spellId: number; talentId: number; rank: number }>;
	professions: string[];
}

export class BadRequestError extends Error {}

// Accepts `/character/<region>/<realm>/<name>`, where the name is `First-Last` or a single name.
export const parseCharacterPath = (pathname: string): CharacterRequest => {
	const match = pathname.match(/^\/character\/([^/]+)\/([^/]+)\/([^/]+)\/?$/);
	if (!match) throw new BadRequestError('Expected /character/<region>/<realm>/<name>');

	const [region, realm, name] = match.slice(1).map(segment => decodeURIComponent(segment).trim());
	if (!(REGIONS as readonly string[]).includes(region.toLowerCase())) throw new BadRequestError(`Unknown region '${region}'`);

	const realmSlug = slugify(realm);
	if (!realmSlug) throw new BadRequestError('Missing realm');

	// Name parts hold no hyphens or spaces in game, so either separates first from last.
	const nameParts = name
		.split(/[\s-]+/)
		.filter(part => part)
		.map(part => part.toLowerCase());
	if (nameParts.length < 1 || nameParts.length > 2) throw new BadRequestError(`Cannot read a character name from '${name}'`);

	return { region: region.toLowerCase() as Region, realmSlug, nameParts };
};

// Realm slugs are lowercase, with spaces as hyphens and apostrophes dropped ("Mirage Raceway" → "mirage-raceway").
export const slugify = (realm: string): string =>
	realm
		.toLowerCase()
		.replace(/['’]/g, '')
		.trim()
		.replace(/\s+/g, '-');

// How Blizzard joins a two-part name in the profile path is not documented, so the separator is a Worker var.
export const characterProfilePath = ({ realmSlug, nameParts }: CharacterRequest, nameSeparator: string): string =>
	`/profile/wow/character/${encodeURIComponent(realmSlug)}/${encodeURIComponent(nameParts.join(nameSeparator))}`;

// The Profile API fields this Worker reads, as the Classic flavours return them.
interface Ref {
	id: number;
	name?: string;
}

export interface ProfileSummary {
	name: string;
	level: number;
	character_class: Ref;
	race: Ref;
	realm: { slug: string; name?: string };
}

export interface ProfileEquipment {
	equipped_items?: Array<{
		item: Ref;
		slot: { type: string };
		enchantments?: Array<{ enchantment_id: number; enchantment_slot?: { id: number; type?: string } }>;
	}>;
}

export interface ProfileSpecializations {
	specialization_groups?: Array<{
		is_active?: boolean;
		specializations?: Array<{
			talents?: Array<{ talent?: Ref; spell_tooltip?: { spell?: Ref }; talent_rank?: number }>;
		}>;
	}>;
}

export interface ProfileProfessions {
	primaries?: Array<{ profession: Ref }>;
}

// The permanent enchant (enchanting, armor kits, ...) sits in enchantment slot 0; other slots hold temporary enchants.
const PERMANENT_ENCHANT_SLOT = 0;

export const trimCharacter = (
	summary: ProfileSummary,
	equipment: ProfileEquipment,
	specializations: ProfileSpecializations,
	professions: ProfileProfessions | null,
): ArmoryCharacter => {
	const groups = specializations.specialization_groups ?? [];
	const activeGroup = groups.find(group => group.is_active) ?? groups[0];

	return {
		name: summary.name,
		realm: summary.realm.name ?? summary.realm.slug,
		level: summary.level,
		classId: summary.character_class.id,
		raceName: summary.race.name ?? '',
		items: (equipment.equipped_items ?? []).map(item => {
			const enchant = item.enchantments?.find(e => (e.enchantment_slot?.id ?? PERMANENT_ENCHANT_SLOT) == PERMANENT_ENCHANT_SLOT)?.enchantment_id;
			return { slot: item.slot.type, id: item.item.id, ...(enchant ? { enchant } : {}) };
		}),
		talents: (activeGroup?.specializations ?? []).flatMap(spec =>
			(spec.talents ?? []).map(talent => ({
				spellId: talent.spell_tooltip?.spell?.id ?? 0,
				talentId: talent.talent?.id ?? 0,
				rank: talent.talent_rank ?? 1,
			})),
		),
		professions: (professions?.primaries ?? []).map(p => p.profession.name ?? '').filter(name => name),
	};
};
