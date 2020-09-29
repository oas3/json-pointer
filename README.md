# JavaScript Object Notation (JSON) Pointer
JSON Pointer defines a string syntax for identifying a specific value within a JavaScript 
Object Notation (JSON) document.

---
A JSON Pointer is a Unicode string (see [RFC4627, Section 3](https://tools.ietf.org/html/rfc4627#section-3)) containing a sequence of 
zero or more reference tokens, each prefixed by a `/` (`%x2F`) character.

Source: [RFC6901](https://tools.ietf.org/html/rfc6901)

### ABNF Representation
```abnf
json-pointer    = *( "/" reference-token )
reference-token = *( unescaped / escaped )
unescaped       = %x00-2E / %x30-7D / %x7F-10FFFF
escaped         = "~" ( "0" / "1" )

array-index     = %x30 / ( %x31-39 *(%x30-39) )
```
