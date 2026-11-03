(def_declaration
  value: (_) @function.inside) @function.around

(when_block) @function.around

(section) @class.around

(row_section) @class.around

(group_section) @class.around

(comment)+ @comment.around

(note)+ @comment.around
