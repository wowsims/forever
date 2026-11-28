package tooltip

// What the conditions on a spell the player knows ($?s) and on an aura or passive the player has ($?a, $?p) read
// off the player, where no talent rank answers them. A spell or aura neither map holds is one no player has: a
// rune, a buff the sim does not model, an item's passive.
type Conditions struct {
	Known map[int64]Expr // $?s
	Auras map[int64]Expr // $?a and $?p
}

type conditionProvider interface {
	GetConditions() *Conditions
}

// The player's expression for a condition on a spell, an aura or a passive, or false where the maps hold none. A
// provider with no conditions answers not ok and decides the condition itself.
func (c *compiler) playerCondition(kind string, spellId int64) (symbol, bool) {
	p, ok := c.provider.(conditionProvider)
	if !ok || p.GetConditions() == nil {
		return symbol{}, false
	}
	var known map[int64]Expr
	switch kind {
	case "s":
		known = p.GetConditions().Known
	case "a", "p":
		known = p.GetConditions().Auras
	default:
		return symbol{}, false
	}
	if expr, ok := known[spellId]; ok {
		return symbol{expr: expr}, true
	}
	return constant(0), true
}

// True for a player with a point in the talent.
func TalentTaken(talent int64) Expr {
	return map[string]Expr{"t": talent}
}

// True for a player of one of the classes, by the game's class ids, which the UI's Class numbers too.
func PlayerOfClass(classes ...int32) Expr {
	return anyOf("CLASS", classes)
}

// True for a player who starts the fight in one of the forms or stances, by their ids in the game files.
func PlayerInForm(forms ...int32) Expr {
	return anyOf("FORM", forms)
}

func anyOf(name string, values []int32) Expr {
	var expr Expr
	for _, v := range values {
		is := exprOp("=", map[string]Expr{"v": name}, float64(v))
		if expr == nil {
			expr = is
		} else {
			expr = exprOp("or", expr, is)
		}
	}
	if expr == nil {
		return 0.0
	}
	return expr
}
