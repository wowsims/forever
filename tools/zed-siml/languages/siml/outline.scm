(section
  name: (section_name) @name) @item

(row_section
  name: (section_name) @name) @item

(group_section
  "group" @context
  name: (string) @name) @item

(var_declaration
  "var" @context
  name: (variable) @name) @item

(def_declaration
  "def" @context
  name: (identifier) @name) @item

(sequence_step
  "sequence" @context
  name: (string) @name) @item

(when_block
  "when" @context
  condition: (_) @name) @item
