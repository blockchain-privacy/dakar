# Attributions

Attributions allow linking external information to addresses. This could be for example ownership information or information where this address has been used. In the address cluster view of the address page, attributions of addresses linked to a cluster can be viewed.

## Import

Import address attributions by uploading a CSV-file.
The file must have five columns (`address`,
`tag`,`description`,`source` and
`category`). The fields `address`
and `tag` are mandatory, the rest are optional.
The file may contain at maximum 1 000 attributions.

The example below would create 5 attributions, linking to 4 addresses.

```text
address;tag;description;source;category
address_1;darknet-address;;;
address_2;twitter-@josh;Josh Noname;https://twitter.com/josh;social media
address_3;facebook-some-user-name;;;social media
address_3;case-123;;;
address_4;exchange-Bitfinex;;;
```
