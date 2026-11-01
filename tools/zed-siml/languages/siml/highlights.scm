; Zed applies the last pattern that matches a node, so the general shapes come first. The language server's semantic
; tokens colour what only the schema and the name table know on top: a spell from an aura, an item, an enum value.

(identifier) @constant

(number) @number

(negative_number) @number

(string) @string

(raw) @string.special

(boolean) @boolean

(none) @constant.builtin

(variable) @variable

(placeholder) @variable.special

(comment) @comment

(note) @comment.doc

[
  "siml"
  "spec"
  "talents"
  "use"
  "var"
  "def"
  "with"
  "when"
  "if"
  "on"
  "until"
  "recast"
  "recheck"
  "for"
  "cast"
  "channel"
  "wait"
  "move"
  "group"
  "schedule"
  "sequence"
] @keyword

[
  "and"
  "or"
  "not"
] @keyword.operator

[
  "item"
  "other"
] @namespace

[
  "=="
  "!="
  ">="
  ">"
  "<="
  "<"
  "+"
  "-"
  "*"
  "/"
  "="
] @operator

[
  "("
  ")"
  "["
  "]"
] @punctuation.bracket

[
  ","
  "."
  ":"
] @punctuation.delimiter

"#" @punctuation.special

; Headers and the top lines.
(section_name) @title

(group_section
  name: (string) @title)

(spec
  spec: (identifier) @type)

; Defs.
(def_declaration
  name: (identifier) @function)

(parameters
  (identifier) @variable.parameter)

; Kinds and defs: called with arguments, written bare as a step, or bare where a value goes.
(call
  name: (identifier) @function)

(row
  action: (identifier) @function)

(action
  step: (identifier) @function)

(binary_expression
  left: (identifier) @function)

(binary_expression
  right: (identifier) @function)

(not_expression
  value: (identifier) @function)

(parenthesized_expression
  (identifier) @function)

(row
  condition: (identifier) @function)

(action
  condition: (identifier) @function)

(when_block
  condition: (identifier) @function)

(var_declaration
  value: (identifier) @function)

(def_declaration
  value: (identifier) @function)

(variable_binding
  value: (identifier) @function)

(placeholder_binding
  value: (identifier) @function)

(wait_step
  value: (identifier) @function)

; Accessors (`target.aura(x).is_active`) and unit-prefixed kinds (`self.unit_distance`): the unit, the head and the
; property.
(member
  property: (identifier) @function.method)

(member
  object: (identifier) @variable.special)

(member
  object: (member
    property: (identifier) @function
    arguments: (arguments)))

(cast_step
  target: (identifier) @variable.special)

(channel_step
  target: (identifier) @variable.special)

; Fields: an entry's, a flag's, a message's and a named argument's.
(entry
  name: (identifier) @property)

(entry
  message: (call
    name: (identifier) @property))

(flag
  name: (identifier) @property)

(named_argument
  name: (identifier) @property)
