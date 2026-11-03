("(" @open ")" @close)

("[" @open "]" @close)

(raw
  "`" @open
  "`" @close
  (#set! rainbow.exclude))
