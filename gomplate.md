
# Encoding

## `Encode`

Encode data as a Base64 string. Specifically, this uses the standard Base64 encoding as defined in [RFC4648 &sect;4](https://tools.ietf.org/html/rfc4648#section-4) (and _not_ the URL-safe encoding).

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The data to encode. Can be a string, a byte array, or a buffer. Other types will be converted to strings first. |

Examples

```console
'{{ base64.Encode "hello world" }}'
aGVsbG8gd29ybGQ=
```
```console
'{{ "hello world" | base64.Encode }}'
aGVsbG8gd29ybGQ=
```

## `Decode`

Decode a Base64 string. This supports both standard ([RFC4648 &sect;4](https://tools.ietf.org/html/rfc4648#section-4)) and URL-safe ([RFC4648 &sect;5](https://tools.ietf.org/html/rfc4648#section-5)) encodings.

This function outputs the data as a string, so it may not be appropriate for decoding binary data. Use [`base64.DecodeBytes`](#base64.DecodeBytes)
for binary data.

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The base64 string to decode |

Examples

```console
'{{ base64.Decode "aGVsbG8gd29ybGQ=" }}'
hello world
```
```console
'{{ "aGVsbG8gd29ybGQ=" | base64.Decode }}'
hello world
```

## `DecodeBytes`

Decode a Base64 string. This supports both standard ([RFC4648 &sect;4](https://tools.ietf.org/html/rfc4648#section-4)) and URL-safe ([RFC4648 &sect;5](https://tools.ietf.org/html/rfc4648#section-5)) encodings.

This function outputs the data as a byte array, so it's most useful for outputting binary data that will be processed further. Use [`base64.Decode`](#base64.Decode) to output a plain string.

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The base64 string to decode |

Examples

```console
'{{ base64.DecodeBytes "aGVsbG8gd29ybGQ=" }}'
[104 101 108 108 111 32 119 111 114 108 100]
```
```console
'{{ "aGVsbG8gd29ybGQ=" | base64.DecodeBytes | conv.ToString }}'
hello world
```
---

These functions help manipulate and query collections of data, like lists (slices, or arrays) and maps (dictionaries).

#### Implementation Note
For the functions that return an array, a Go `[]interface{}` is returned, regardless of whether or not the input was a different type.

# `Collection`

## `dict`

Dict is a convenience function that creates a map with string keys. Provide arguments as key/value pairs. If an odd number of arguments is provided, the last is used as the key, and an empty string is set as the value.

All keys are converted to strings.

This function is equivalent to [Sprig's `dict`](http://masterminds.github.io/sprig/dicts.html#dict) function, as used in [Helm templates](https://docs.helm.sh/chart_template_guide#template-functions-and-pipelines).

For creating more complex maps, see [`data.JSON`](../data/#data-json) or [`data.YAML`](../data/#data-yaml).

For creating arrays, see [`coll.Slice`](#coll-slice).

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ The key/value pairs |

Examples

```console
'{{ coll.Dict "name" "Frank" "age" 42 | data.ToYAML }}'
age: 42
name: Frank
'{{ dict 1 2 3 | toJSON }}'
{"1":2,"3":""}
```
```console
$ cat <<EOF| gomplate
{{ define "T1" }}Hello {{ .thing }}!{{ end -}}
{{ template "T1" (dict "thing" "world")}}
{{ template "T1" (dict "thing" "everybody")}}
EOF
Hello world!
Hello everybody!
```

## `slice`

Creates a slice (like an array or list). Useful when needing to `range` over a bunch of variables.

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the elements of the slice |

Examples

```console
'{{ range slice "Bart" "Lisa" "Maggie" }}Hello, {{ . }}{{ end }}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```


## `has`

Reports whether a given object has a property with the given key, or whether a given array/slice contains the given value. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The object or list to search |
| `item` | _(required)_ The item to search for |

Examples

```console
'{{ $l := slice "foo" "bar" "baz" }}there is {{ if has $l "bar" }}a{{else}}no{{end}} bar'
there is a bar
```
```console
$ export DATA='{"foo": "bar"}'
$ '{{ $o := data.JSON (getenv "DATA") -}}
{{ if (has $o "foo") }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}'
bar
```
```console
$ export DATA='{"baz": "qux"}'
$ '{{ $o := data.JSON (getenv "DATA") -}}
{{ if (has $o "foo") }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}'
THERE IS NO FOO
```


## `jsonpath`

Extracts portions of an input object or list using a [JSONPath][] expression.

Any object or list may be used as input. The output depends somewhat on the expression; if multiple items are matched, an array is returned.

JSONPath expressions can be validated at https://jsonpath.com

[JSONPath]: https://goessner.net/articles/JsonPath

Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The JSONPath expression |
| `in` | _(required)_ The object or list to query |

Examples

```console
$ '{{ .books | jsonpath `$..works[?( @.edition_count > 400 )].title` }}' -c books=https://openlibrary.org/subjects/fantasy.json
[Alice's Adventures in Wonderland Gulliver's Travels]
```


## `jq`

Filters an input object or list using the [jq](https://stedolan.github.io/jq/) language, as implemented by [gojq](https://github.com/itchyny/gojq).

Any JSON datatype may be used as input (NOTE: strings are not JSON-parsed but passed in as is). If the expression results in multiple items (no matter if streamed or as an array) they are wrapped in an array. Otherwise a single item is returned (even if resulting in an array with a single contained element).

JQ filter expressions can be tested at https://jqplay.org/

See also:

- [jq manual](https://stedolan.github.io/jq/manual/)
- [gojq differences to jq](https://github.com/itchyny/gojq#difference-to-jq)


Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The JQ expression |
| `in` | _(required)_ The object or list to query |

Examples

```console
$ '{{ .books | jq `[.works[]|{"title":.title,"authors":[.authors[].name],"published":.first_publish_year}][0]` }}' -c books=https://openlibrary.org/subjects/fantasy.json

map[authors:[Lewis Carroll] published:1865 title:Alice's Adventures in Wonderland]
```


## `keys`

Return a list of keys in one or more maps.

The keys will be ordered first by map position (if multiple maps are given), then alphabetically.

See also [`coll.Values`](#coll-values).


Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the maps |

Examples

```console
$ '{{ coll.Keys (dict "foo" 1 "bar" 2) }}'
[bar foo]
$ '{{ $map1 := dict "foo" 1 "bar" 2 -}}{{ $map2 := dict "baz" 3 "qux" 4 -}}{{ coll.Keys $map1 $map2 }}'
[bar foo baz qux]
```


## `values`

Return a list of values in one or more maps.

The values will be ordered first by map position (if multiple maps are given), then alphabetically by key.

See also [`coll.Keys`](#coll-keys).


Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the maps |

Examples

```console
$ '{{ coll.Values (dict "foo" 1 "bar" 2) }}'
[2 1]
$ '{{ $map1 := dict "foo" 1 "bar" 2 -}}{{ $map2 := dict "baz" 3 "qux" 4 -}}{{ coll.Values $map1 $map2 }}'
[2 1 3 4]
```


## `append`

Append a value to the end of a list.

_Note that this function does not change the given list; it always produces a new one._

See also [`coll.Prepend`](#coll-prepend).

Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to add |
| `list...` | _(required)_ the slice or array to append to |

Examples

```console
$ '{{ slice 1 1 2 3 | append 5 }}'
[1 1 2 3 5]
```


## `prepend`

Prepend a value to the beginning of a list.

_Note that this function does not change the given list; it always produces a new one._

See also [`coll.Append`](#coll-append).


Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to add |
| `list...` | _(required)_ the slice or array to prepend to |

Examples

```console
$ '{{ slice 4 3 2 1 | prepend 5 }}'
[5 4 3 2 1]
```

## `uniq`

Remove any duplicate values from the list, without changing order.

_Note that this function does not change the given list; it always produces a new one._


Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ the input list |

Examples

```console
$ '{{ slice 1 2 3 2 3 4 1 5 | uniq }}'
[1 2 3 4 5]
```


## `flatten`

Flatten a nested list. Defaults to completely flattening all nested lists, but can be limited with `depth`.

_Note that this function does not change the given list; it always produces a new one._


Arguments

| name | description |
|------|-------------|
| `depth` | _(optional)_ maximum depth of nested lists to flatten. Omit or set to `-1` for infinite depth. |
| `list` | _(required)_ the input list |

Examples

```console
$ '{{ "[[1,2],[],[[3,4],[[[5],6],7]]]" | jsonArray | flatten }}'
[1 2 3 4 5 6 7]
```
```console
$ '{{ coll.Flatten 2 ("[[1,2],[],[[3,4],[[[5],6],7]]]" | jsonArray) }}'
[1 2 3 4 [[5] 6] 7]
```


## `reverse`

Reverse a list.

_Note that this function does not change the given list; it always produces a new one._


Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ the list to reverse |

Examples

```console
$ '{{ slice 4 3 2 1 | reverse }}'
[1 2 3 4]
```


## `sort`

Sort a given list. Uses the natural sort order if possible. For inputs that are not sortable (either because the elements are of different types, or of an un-sortable type), the input will simply be returned, unmodified.

Maps and structs can be sorted by a named key.

_Note that this function does not modify the input._


Arguments

| name | description |
|------|-------------|
| `key` | _(optional)_ the key to sort by, for lists of maps or structs |
| `list` | _(required)_ the slice or array to sort |

Examples

```console
$ '{{ slice "foo" "bar" "baz" | coll.Sort }}'
[bar baz foo]
```
```console
$ '{{ sort (slice 3 4 1 2 5) }}'
[1 2 3 4 5]
```


## `merge`

Merge maps together by overriding src with dst. In other words, the src map can be configured the "default" map, whereas the dst
map can be configured the "overrides". Many source maps can be provided. Precedence is in left-to-right order.

_Note that this function does not modify the input._


Arguments

| name | description |
|------|-------------|
| `dst` | _(required)_ the map to merge _into_ |
| `srcs...` | _(required)_ the map (or maps) to merge _from_ |

Examples

```console
$ '{{ $default := dict "foo" 1 "bar" 2}}
{{ $config := dict "foo" 8 }}
{{ merge $config $default }}'
map[bar:2 foo:8]
```
```console
$ '{{ $dst := dict "foo" 1 "bar" 2 }}
{{ $src1 := dict "foo" 8 "baz" 4 }}
{{ $src2 := dict "foo" 3 "bar" 5 }}
{{ coll.Merge $dst $src1 $src2 }}'
map[foo:1 bar:5 baz:4]
```


Given a map, returns a new map with any entries that have the given keys.

All keys are converted to strings.

This is the inverse of [`coll.Omit`](#coll-omit).

_Note that this function does not modify the input._

Arguments

| name | description |
|------|-------------|
| `keys...` | _(required)_ the keys to match |
| `map` | _(required)_ the map to pick from |

Examples

```console
$ '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Pick "foo" "baz" $data }}'
map[baz:3 foo:1]
```

Given a map, returns a new map without any entries that have the given keys.

All keys are converted to strings.

This is the inverse of [`coll.Pic`](#coll-pick).

_Note that this function does not modify the input._

Arguments

| name | description |
|------|-------------|
| `keys...` | _(required)_ the keys to match |
| `map` | _(required)_ the map to omit from |

Examples

```console
$ '{{ $data := dict "foo" 1 "bar" 2 "baz" 3 }}
{{ coll.Omit "foo" "baz" $data }}'
map[bar:2]
```

# Convert

## `bool`

**Note:** See also [`conv.ToBool`](#conv-tobool) for a more flexible variant.

Converts a true-ish string to a boolean. Can be used to simplify conditional statements based on environment variables or other text input.


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

Examples

```console
$ FOO=true 
$ {{if bool (getenv "FOO")}}foo{{else}}bar{{end}}
foo
```

## `default`

Provides a default value given an empty input. Empty inputs are `0` for numeric types, `""` for strings, `false` for booleans, empty arrays/maps, and `nil`.

Note that this will not provide a default for the case where the input is undefined (i.e. referencing things like `.foo` where there is no `foo` field of `.`), but [`conv.Has`](#conv-has) can be used for that.

Arguments

| name | description |
|------|-------------|
| `default` | _(required)_ the default value |
| `in` | _(required)_ the input |

Examples

```console
$ '{{ "" | default "foo" }} {{ "bar" | default "baz" }}'
foo bar
```

## `dict`

Dict is a convenience function that creates a map with string keys. Provide arguments as key/value pairs. If an odd number of arguments
is provided, the last is used as the key, and an empty string is set as the value.

All keys are converted to strings.

This function is equivalent to [Sprig's `dict`](http://masterminds.github.io/sprig/dicts.html#dict) function, as used in [Helm templates](https://docs.helm.sh/chart_template_guide#template-functions-and-pipelines).

For creating more complex maps, see [`data.JSON`](../data/#data-json) or [`data.YAML`](../data/#data-yaml).

For creating arrays, see [`conv.Slice`](#conv-slice).

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ The key/value pairs |

Examples

```console
$ '{{ conv.Dict "name" "Frank" "age" 42 | data.ToYAML }}'
age: 42
name: Frank
$ '{{ dict 1 2 3 | toJSON }}'
{"1":2,"3":""}
```

## `slice`

Creates a slice (like an array or list). Useful when needing to `range` over a bunch of variables.


Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the elements of the slice |

Examples

```console
$ '{{ range slice "Bart" "Lisa" "Maggie" }}Hello, {{ . }}{{ end }}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `has`

Reports whether a given object has a property with the given key, or whether a given array/slice contains the given value. Can be used with `if` to prevent the template from trying to access a non-existent property in an object.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The object or list to search |
| `item` | _(required)_ The item to search for |

Examples

```console
$ '{{ $l := slice "foo" "bar" "baz" }}there is {{ if has $l "bar" }}a{{else}}no{{end}} bar'
there is a bar
```
```console
$ export DATA='{"foo": "bar"}'
$ '{{ $o := data.JSON (getenv "DATA") -}}
{{ if (has $o "foo") }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}'
bar
```
```console
$ export DATA='{"baz": "qux"}'
$ '{{ $o := data.JSON (getenv "DATA") -}}
{{ if (has $o "foo") }}{{ $o.foo }}{{ else }}THERE IS NO FOO{{ end }}'
THERE IS NO FOO
```


## `join`

Concatenates the elements of an array to create a string. The separator string `sep` is placed between elements in the resulting string.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the array or slice |
| `sep` | _(required)_ the separator |

Examples

```console
$ '{{ $a := slice 1 2 3 }}{{ join $a "-" }}'
1-2-3
```


## `urlParse`

Parses a string as a URL for later use. Equivalent to [url.Parse](https://golang.org/pkg/net/url/#Parse)

Any of `url.URL`'s methods can be called on the result.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the URL string to parse |

Examples

_`input.tmpl`:_
```
{{ $u := conv.URL "https://example.com:443/foo/bar" }}
The scheme is {{ $u.Scheme }}
The host is {{ $u.Host }}
The path is {{ $u.Path }}
```

```console
$ gomplate < input.tmpl
The scheme is https
The host is example.com:443
The path is /foo/bar
```
_Call `Redacted` to hide the password in the output:_
```
$ '{{ (conv.URL "https://user:supersecret@example.com").Redacted }}'
https://user:xxxxx@example.com
```

## `ParseInt`

_**Note:**_ See [`conv.ToInt64`](#conv-toint64) instead for a simpler and more flexible variant of this function.

Parses a string as an int64. Equivalent to [strconv.ParseInt](https://golang.org/pkg/strconv/#ParseInt)


Examples

```console
$ HEXVAL=7C0 
$ {{ $val := conv.ParseInt (getenv "HEXVAL") 16 32 }}
$ The value in decimal is {{ $val }}

The value in decimal is 1984
```

## `ParseFloat`

_**Note:**_ See [`conv.ToFloat`](#conv-tofloat) instead for a simpler and more flexible variant of this function.

Parses a string as an float64 for later use. Equivalent to [strconv.ParseFloat](https://golang.org/pkg/strconv/#ParseFloat)

Examples

_`input.tmpl`:_
```
{{ $pi := conv.ParseFloat (getenv "PI") 64 }}
{{- if (gt $pi 3.0) -}}
pi is greater than 3
{{- end }}
```

```console
{{ $pi := conv.ParseFloat (getenv "PI") 64 }}
{{- if (gt $pi 3.0) -}}
pi is greater than 3
{{- end }}
PI=3.14159265359

pi is greater than 3
```

## `ParseUint`

Parses a string as an uint64 for later use. Equivalent to [strconv.ParseUint](https://golang.org/pkg/strconv/#ParseUint)

Examples

```console
{{ conv.ParseInt (getenv "BIG") 16 64 }} is max int64
{{ conv.ParseUint (getenv "BIG") 16 64 }} is max uint64
$ BIG=FFFFFFFFFFFFFFFF

9223372036854775807 is max int64
18446744073709551615 is max uint64
```

## `.Atoi`

_**Note:**_ See [`conv.ToInt`](#conv-toint) and [`conv.ToInt64`](#conv-toint64) instead for simpler and more flexible variants of this function.

Parses a string as an int for later use. Equivalent to [strconv.Atoi](https://golang.org/pkg/strconv/#Atoi)


Examples

```console
$ NUMBER=21
{{ $number := conv.Atoi (getenv "NUMBER") }}
{{- if (gt $number 5) -}}
The number is greater than 5
{{- else -}}
The number is less than 5
{{- end }}

The number is greater than 5
```

## `ToBool`

Converts the input to a boolean value. Possible `true` values are: `1` or the strings `"t"`, `"true"`, or `"yes"` (any capitalizations). All other values are considered `false`.


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input to convert |

Examples

```console
$ '{{ conv.ToBool "yes" }} {{ conv.ToBool true }} {{ conv.ToBool "0x01" }}'
true true true
$ '{{ conv.ToBool false }} {{ conv.ToBool "blah" }} {{ conv.ToBool 0 }}'
false false false
```

## `ToBools`

Converts a list of inputs to an array of boolean values. Possible `true` values are: `1` or the strings `"t"`, `"true"`, or `"yes"` (any capitalizations). All other values are considered `false`.


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input array to convert |

Examples

```console
$ '{{ conv.ToBools "yes" true "0x01" }}'
[true true true]
$ '{{ conv.ToBools false "blah" 0 }}'
[false false false]
```

## `ToInt64`

Converts the input to an `int64` (64-bit signed integer).

This function attempts to convert most types of input (strings, numbers, and booleans), but behaviour when the input can not be converted is undefined and subject to change. Unconvertable inputs may result in errors, or `0` or `-1`.

Floating-point numbers (with decimal points) are truncated.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

Examples

```console
$ '{{conv.ToInt64 "9223372036854775807"}}'
9223372036854775807
```
```console
$ '{{conv.ToInt64 "0x42"}}'
66
```
```console
$ '{{conv.ToInt64 true }}'
1
```

## `ToInt`

Converts the input to an `int` (signed integer, 32- or 64-bit depending on platform). This is similar to [`conv.ToInt64`](#conv-toint64) on 64-bit platforms, but is useful when input to another function must be provided as an `int`.

On 32-bit systems, given a number that is too large to fit in an `int`, the result is `-1`. This is done to protect against
[CWE-190](https://cwe.mitre.org/data/definitions/190.html) and [CWE-681](https://cwe.mitre.org/data/definitions/681.html).

See also [`conv.ToInt64`](#conv-toint64).

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

Examples

```console
$ '{{conv.ToInt "9223372036854775807"}}'
9223372036854775807
```
```console
$ '{{conv.ToInt "0x42"}}'
66
```
```console
$ '{{conv.ToInt true }}'
1
```

## `ToInt64s`

Converts the inputs to an array of `int64`s.

This delegates to [`conv.ToInt64`](#conv-toint64) for each input argument.

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

Examples

```console
'{{ conv.ToInt64s true 0x42 "123,456.99" "1.2345e+3"}}'
[1 66 123456 1234]
```

## `ToInts`

Converts the inputs to an array of `int`s.

This delegates to [`conv.ToInt`](#conv-toint) for each input argument.

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

Examples

```console
'{{ conv.ToInts true 0x42 "123,456.99" "1.2345e+3"}}'
[1 66 123456 1234]
```

## `ToFloat64`

Converts the input to a `float64`.

This function attempts to convert most types of input (strings, numbers, and booleans), but behaviour when the input can not be converted is undefined and subject to change. Unconvertable inputs may result in errors, or `0` or `-1`.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

Examples

```console
$ '{{ conv.ToFloat64 "8.233e-1"}}'
0.8233
$ '{{ conv.ToFloat64 "9,000.09"}}'
9000.09
```

## `ToFloat64s`

Converts the inputs to an array of `float64`s.

This delegates to [`conv.ToFloat64`](#conv-tofloat64) for each input argument.

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

Examples

```console
$ '{{ conv.ToFloat64s true 0x42 "123,456.99" "1.2345e+3"}}'
[1 66 123456.99 1234.5]
```

## `ToString`

Converts the input (of any type) to a `string`.

The input will always be represented in _some_ way.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the value to convert |

Examples

```console
$ '{{ conv.ToString 0xFF }}'
255
$ '{{ dict "foo" "bar" | conv.ToString}}'
map[foo:bar]
$ '{{ conv.ToString nil }}'
nil
```

## `ToStrings`

Converts the inputs (of any type) to an array of `string`s

This delegates to [`conv.ToString`](#conv-tostring) for each input argument.

Arguments

| name | description |
|------|-------------|
| `in...` | _(required)_ the inputs to be converted |

Examples

```console
$ '{{ conv.ToStrings nil 42 true 0xF (slice 1 2 3) }}'
[nil 42 true 15 [1 2 3]]
```


# `Cryptography`

## `crypto.SHA1`, `crypto.SHA224`, `crypto.SHA256`, `crypto.SHA384`, `crypto.SHA512`, `crypto.SHA512_224`, `crypto.SHA512_256`

Compute a checksum with a SHA-1 or SHA-2 algorithm as defined in [RFC 3174](https://tools.ietf.org/html/rfc3174) (SHA-1) and [FIPS 180-4](http://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf) (SHA-2).

These functions output the binary result as a hexadecimal string.

_Warning: SHA-1 is cryptographically broken and should not be used for secure applications._

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the data to hash - can be binary data or text |

Examples

```console
$ '{{ crypto.SHA1 "foo" }}'
f1d2d2f924e986ac86fdf7b36c94bcdf32beec15
```
```console
$ '{{ crypto.SHA512 "bar" }}'
cc06808cbbee0510331aa97974132e8dc296aeb795be229d064bae784b0a87a5cf4281d82e8c99271b75db2148f08a026c1a60ed9cabdb8cac6d24242dac4063
```

## `crypto.SHA1Bytes`, `crypto.SHA224Bytes`, `crypto.SHA256Bytes`, `crypto.SHA384Bytes`, `crypto.SHA512Bytes`, `crypto.SHA512_224Bytes`, `crypto.SHA512_256Bytes`

Compute a checksum with a SHA-1 or SHA-2 algorithm as defined in [RFC 3174](https://tools.ietf.org/html/rfc3174) (SHA-1) and [FIPS 180-4](http://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf) (SHA-2).

These functions output the raw binary result, suitable for piping to other functions.

_Warning: SHA-1 is cryptographically broken and should not be used for secure applications._

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the data to hash - can be binary data or text |

Examples

```console
$ '{{ crypto.SHA256Bytes "foo" | base64.Encode }}'
LCa0a2j/xo/5m0U8HTBBNBNCLXBkg7+g+YpeiGJm564=
```

## `include`

Includes the content of a given datasource (provided by the [`--datasource/-d`](../../usage/#datasource-d) argument).

This is similar to [`datasource`](#datasource), except that the data is not parsed. There is no restriction on the type of data included, except that it should be textual.

Arguments

| name | description |
|------|-------------|
| `alias` | _(required)_ the datasource alias, as provided by [`--datasource/-d`](../../usage/#datasource-d) |
| `subpath` | _(optional)_ the subpath to use, if supported by the datasource |

Examples

_`person.json`:_
```json
{ "name": "Dave" }
```

_`input.tmpl`:_
```go
{
  "people": [
    {{ include "person" }}
  ]
}
```
```
{
  "people": [
    { "name": "Dave" }
  ]
}
```

# `Data`

## `json`

Converts a JSON string into an object. Works for JSON Objects, but will also parse JSON Arrays. Will not parse other valid JSON types.

For more explict JSON Array support, see [`data.JSONArray`](#data-jsonarray).

#### Encrypted JSON support (EJSON)

If the input is in the [EJSON](https://github.com/Shopify/ejson) format (i.e. has a `_public_key` field), this function will attempt to decrypt the document first. A private key must be provided by one of these methods:

- set the `EJSON_KEY` environment variable to the private key's value
- set the `EJSON_KEY_FILE` environment variable to the path to a file containing the private key
- set the `EJSON_KEYDIR` environment variable to the path to a directory containing private keys (filename must be the public key), just like [`ejson decrypt`'s `--keydir`](https://github.com/Shopify/ejson/blob/master/man/man1/ejson.1.ronn) flag. Defaults to `/opt/ejson/keys`.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

Examples

```console
$ export FOO='{"hello":"world"}'
$ Hello {{ (getenv "FOO" | json).hello }}
Hello world
```


## `jsonArray`

Converts a JSON string into a slice. Only works for JSON Arrays.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

Examples

```console
$ export FOO='[ "you", "world" ]'
Hello {{ index (getenv "FOO" | jsonArray) 1 }}
Hello world
```


## `yaml`

Converts a YAML string into an object. Works for YAML Objects but will also parse YAML Arrays. This can be used to access properties of YAML objects.

For more explict YAML Array support, see [`data.JSONArray`](#data-yamlarray).

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

Examples


```console
$ export FOO='hello: world'
$ Hello {{ (getenv "FOO" | yaml).hello }}

Hello world
```


## `yamlArray`

Converts a YAML string into a slice. Only works for YAML Arrays.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ the input string |

Examples

```console
$ export FOO='[ "you", "world" ]'
$ Hello {{ index (getenv "FOO" | yamlArray) 1 }}
Hello world
```


## `toml`

Converts a [TOML](https://github.com/toml-lang/toml) document into an object. This can be used to access properties of TOML documents.

Compatible with [TOML v0.4.0](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md).

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the TOML document to parse |

Examples

```console
{{ $t := `[data]
hello = "world"` -}}
Hello {{ (toml $t).hello }}

Hello world
```


## `csv`

Converts a CSV-format string into a 2-dimensional string array.

By default, the [RFC 4180](https://tools.ietf.org/html/rfc4180) format is supported, but any single-character delimiter can be specified.

Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `input` | _(required)_ the CSV-format string to parse |

Examples

_`input.tmpl`:_
```
{{ $c := `C,32
Go,25
COBOL,357` -}}
{{ range ($c | csv) -}}
{{ index . 0 }} has {{ index . 1 }} keywords.
{{ end }}
```

```console
C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords.
```

## `csvByRow`

Converts a CSV-format string into a slice of maps.

By default, the [RFC 4180](https://tools.ietf.org/html/rfc4180) format is supported, but any single-character delimiter can be specified.

Also by default, the first line of the string will be assumed to be the header, but this can be overridden by providing an explicit header, or auto-indexing can be used.

Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `header` | _(optional)_ comma-separated list of column names, set to `""` to get auto-named columns (A-Z), defaults to using the first line of `input` |
| `input` | _(required)_ the CSV-format string to parse |

Examples

_`input.tmpl`:_
```
{{ $c := `lang,keywords
C,32
Go,25
COBOL,357` -}}
{{ range ($c | csvByRow) -}}
{{ .lang }} has {{ .keywords }} keywords.
{{ end }}
```

```console
C has 32 keywords.
Go has 25 keywords.
COBOL has 357 keywords.
```


## `csvByColumn`

Like [`csvByRow`](#csvByRow), except that the data is presented as a columnar (column-oriented) map.

Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `header` | _(optional)_ comma-separated list of column names, set to `""` to get auto-named columns (A-Z), defaults to using the first line of `input` |
| `input` | _(required)_ the CSV-format string to parse |

Examples

_`input.tmpl`:_
```
{{ $c := `C;32
Go;25
COBOL;357` -}}
{{ $langs := ($c | csvByColumn ";" "lang,keywords").lang -}}
{{ range $langs }}{{ . }}
{{ end -}}
```

```console
C
Go
COBOL
```


## `toJSON`

Converts an object to a JSON document. Input objects may be the result of `json`, `yaml`, `jsonArray`, or `yamlArray` functions, or they could be provided by a `datasource`.


Arguments

| name | description |
|------|-------------|
| `obj` | _(required)_ the object to marshal |

Examples

_This is obviously contrived - `json` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | json).foo | toJSON }}
```

```console
{"hello":"world"}
```

## `toJSONPretty`

Converts an object to a pretty-printed (or _indented_) JSON document. Input objects may be the result of functions like `data.JSON`, `data.YAML`, `data.JSONArray`, or `data.YAMLArray` functions, or they could be provided by a [`datasource`](../general/datasource).

The indent string must be provided as an argument.

Arguments

| name | description |
|------|-------------|
| `indent` | _(required)_ the string to use for indentation |
| `obj` | _(required)_ the object to marshal |

Examples

_`input.tmpl`:_
```
{{ `{"hello":"world"}` | data.JSON | data.ToJSONPretty "  " }}
```

```console
{
  "hello": "world"
}
```

## `toYAML`

Converts an object to a YAML document. Input objects may be the result of `data.JSON`, `data.YAML`, `data.JSONArray`, or `data.YAMLArray` functions, or they could be provided by a [`datasource`](../general/datasource).

Arguments

| name | description |
|------|-------------|
| `obj` | _(required)_ the object to marshal |

Examples

_This is obviously contrived - `data.JSON` is used to create an object._

_`input.tmpl`:_
```
{{ (`{"foo":{"hello":"world"}}` | data.JSON).foo | data.ToYAML }}
```

```console
hello: world
```

## `toTOML`

Converts an object to a [TOML](https://github.com/toml-lang/toml) document.

Arguments

| name | description |
|------|-------------|
| `obj` | _(required)_ the object to marshal as a TOML document |

Examples

```console
$ '{{ `{"foo":"bar"}` | data.JSON | data.ToTOML }}'
foo = "bar"
```


## `toCSV`

Converts an object to a CSV document. The input object must be a 2-dimensional array of strings (a `[][]string`). Objects produced by [`data.CSVByRow`](#conv-csvbyrow) and [`data.CSVByColumn`](#conv-csvbycolumn) cannot yet be converted back to CSV documents.

**Note:** With the exception that a custom delimiter can be used, `data.ToCSV` outputs according to the [RFC 4180](https://tools.ietf.org/html/rfc4180) format, which means that line terminators are `CRLF` (Windows format, or `\r\n`). If you require `LF` (UNIX format, or `\n`), the output can be piped through [`strings.ReplaceAll`](../strings/#strings-replaceall) to replace `"\r\n"` with `"\n"`.

Arguments

| name | description |
|------|-------------|
| `delim` | _(optional)_ the (single-character!) field delimiter, defaults to `","` |
| `input` | _(required)_ the object to convert to a CSV |

Examples

_`input.tmpl`:_
```go
{{ $rows := (jsonArray `[["first","second"],["1","2"],["3","4"]]`) -}}
{{ data.ToCSV ";" $rows }}
```

```console
first,second
1,2
3,4
```



# `filepath`

## `Base`

Returns the last element of path. Trailing path separators are removed before extracting the last element. If the path is empty, Base returns `.`. If the path consists entirely of separators, Base returns a single separator.

A wrapper for Go's [`filepath.Base`](https://golang.org/pkg/path/filepath/#Base) function.

Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ filepath.Base "/tmp/foo" }}'
foo
```

## `Clean`

Clean returns the shortest path name equivalent to path by purely lexical processing.

A wrapper for Go's [`filepath.Clean`](https://golang.org/pkg/path/filepath/#Clean) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ filepath.Clean "/tmp//foo/../" }}'
/tmp
```

## `Dir`

Returns all but the last element of path, typically the path's directory.

A wrapper for Go's [`filepath.Dir`](https://golang.org/pkg/path/filepath/#Dir) function.

Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ filepath.Dir "/tmp/foo" }}'
/tmp
```

## `Ext`

Returns the file name extension used by path.

A wrapper for Go's [`filepath.Ext`](https://golang.org/pkg/path/filepath/#Ext) function.

Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ filepath.Ext "/tmp/foo.csv" }}'
.csv
```

## `FromSlash`

Returns the result of replacing each slash (`/`) character in the path with the platform's separator character.

A wrapper for Go's [`filepath.FromSlash`](https://golang.org/pkg/path/filepath/#FromSlash) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ filepath.FromSlash "/foo/bar" }}'
/foo/bar
```

## `IsAbs`

Reports whether the path is absolute.

A wrapper for Go's [`filepath.IsAbs`](https://golang.org/pkg/path/filepath/#IsAbs) function.

Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ the path is {{ if (filepath.IsAbs "/tmp/foo.csv") }}absolute{{else}}relative{{end}} 
the path is absolute

$ the path is {{ if (filepath.IsAbs "../foo.csv") }}absolute{{else}}relative{{end}} 
the path is relative
```

## `Join`

Joins any number of path elements into a single path, adding a separator if necessary.

A wrapper for Go's [`filepath.Join`](https://golang.org/pkg/path/filepath/#Join) function.

Arguments

| name | description |
|------|-------------|
| `elem...` | _(required)_ The path elements to join (0 or more) |

Examples

```console
$ '{{ filepath.Join "/tmp" "foo" "bar" }}'
/tmp/foo/bar
```

## `Match`

Reports whether name matches the shell file name pattern.

A wrapper for Go's [`filepath.Match`](https://golang.org/pkg/path/filepath/#Match) function.


Arguments

| name | description |
|------|-------------|
| `pattern` | _(required)_ The pattern to match on |
| `path` | _(required)_ The path to match |

Examples

```console
$ '{{ filepath.Match "*.csv" "foo.csv" }}'
true
```

## `Rel`

Returns a relative path that is lexically equivalent to targetpath when joined to basepath with an intervening separator.

A wrapper for Go's [`filepath.Rel`](https://golang.org/pkg/path/filepath/#Rel) function.



Arguments

| name | description |
|------|-------------|
| `basepath` | _(required)_ The base path |
| `targetpath` | _(required)_ The target path |

Examples

```console
$ '{{ filepath.Rel "/a" "/a/b/c" }}'
b/c
```

## `Split`

Splits path immediately following the final path separator, separating it into a directory and file name component.

The function returns an array with two values, the first being the diretory, and the second the file.

A wrapper for Go's [`filepath.Split`](https://golang.org/pkg/path/filepath/#Split) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ $p := filepath.Split "/tmp/foo" }}{{ $dir := index $p 0 }}{{ $file := index $p 1 }}dir is {{$dir}}, file is {{$file}}'
dir is /tmp/, file is foo
```

## `ToSlash`

Returns the result of replacing each separator character in path with a slash (`/`) character.

A wrapper for Go's [`filepath.ToSlash`](https://golang.org/pkg/path/filepath/#ToSlash) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ filepath.ToSlash "/foo/bar" }}'
/foo/bar
```

## `VolumeName`

Returns the leading volume name. Given `C:\foo\bar` it returns `C:` on Windows. Given a UNC like `\\host\share\foo` it returns `\\host\share`. On other platforms it returns an empty string.

A wrapper for Go's [`filepath.VolumeName`](https://golang.org/pkg/path/filepath/#VolumeName) function.



Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ 'volume is {{ filepath.VolumeName "C:/foo/bar" }}'
volume is C:
$ 'volume is {{ filepath.VolumeName "/foo/bar" }}'
volume is
```



# `math`

Returns the absolute value of a given number. When the input is an integer, the result will be an `int64`, otherwise it will be a `float64`.



Arguments

| name | description |
|------|-------------|
| `num` | _(required)_ The input number |

Examples

```console
$ '{{ math.Abs -3.5 }} {{ math.Abs 3.5 }} {{ math.Abs -42 }}'
3.5 3.5 42
```


## `add`

Adds all given operators. When one of the inputs is a floating-point number, the result will be a `float64`, otherwise it will be an `int64`.


Arguments

| name | description |
|------|-------------|
| `n...` | _(required)_ The numbers to add together |

Examples

```console
$ '{{ math.Add 1 2 3 4 }} {{ math.Add 1.5 2 3 }}'
10 6.5
```

## `Ceil`

Returns the least integer value greater than or equal to a given floating-point number. This wraps Go's [`math.Ceil`](https://golang.org/pkg/math/#Ceil).

**Note:** the return value of this function is a `float64` so that the special-cases `NaN` and `Inf` can be returned appropriately.


Arguments

| name | description |
|------|-------------|
| `num` | _(required)_ The input number. Will be converted to a `float64`, or `0` if not convertible |

Examples

```console
$ '{{ range (slice 5.1 42 "3.14" "0xFF" "NaN" "Inf" "-0") }}ceil {{ printf "%#v" . }} = {{ math.Ceil . }}{{"\n"}}{{ end }}'
ceil 5.1 = 6
ceil 42 = 42
ceil "3.14" = 4
ceil "0xFF" = 255
ceil "NaN" = NaN
ceil "Inf" = +Inf
ceil "-0" = 0
```


## `div`

Divide the first number by the second. Division by zero is disallowed. The result will be a `float64`.


Arguments

| name | description |
|------|-------------|
| `a` | _(required)_ The divisor |
| `b` | _(required)_ The dividend |

Examples

```console
$ '{{ math.Div 8 2 }} {{ math.Div 3 2 }}'
4 1.5
```

## `Floor`

Returns the greatest integer value less than or equal to a given floating-point number. This wraps Go's [`math.Floor`](https://golang.org/pkg/math/#Floor).

**Note:** the return value of this function is a `float64` so that the special-cases `NaN` and `Inf` can be returned appropriately.


Arguments

| name | description |
|------|-------------|
| `num` | _(required)_ The input number. Will be converted to a `float64`, or `0` if not convertable |

Examples

```console
$ '{{ range (slice 5.1 42 "3.14" "0xFF" "NaN" "Inf" "-0") }}floor {{ printf "%#v" . }} = {{ math.Floor . }}{{"\n"}}{{ end }}'
floor 5.1 = 4
floor 42 = 42
floor "3.14" = 3
floor "0xFF" = 255
floor "NaN" = NaN
floor "Inf" = +Inf
floor "-0" = 0
```

## `IsFloat`

Returns whether or not the given number can be interpreted as a floating-point literal, as defined by the [Go language reference](https://golang.org/ref/spec#Floating-point_literals).

**Note:** If a decimal point is part of the input number, it will be considered a floating-point number, even if the decimal is `0`.


Arguments

| name | description |
|------|-------------|
| `num` | _(required)_ The value to test |

Examples

```console
$ '{{ range (slice 1.0 "-1.0" 5.1 42 "3.14" "foo" "0xFF" "NaN" "Inf" "-0") }}{{ if (math.IsFloat .) }}{{.}} is a float{{"\n"}}{{ end }}{{end}}'
1 is a float
-1.0 is a float
5.1 is a float
3.14 is a float
NaN is a float
Inf is a float
```

## `IsInt`

Returns whether or not the given number is an integer.



Arguments

| name | description |
|------|-------------|
| `num` | _(required)_ The value to test |

Examples

```console
$ '{{ range (slice 1.0 "-1.0" 5.1 42 "3.14" "foo" "0xFF" "NaN" "Inf" "-0") }}{{ if (math.IsInt .) }}{{.}} is an integer{{"\n"}}{{ end }}{{end}}'
42 is an integer
0xFF is an integer
-0 is an integer
```

## `IsNum`

Returns whether the given input is a number. Useful for `if` conditions.


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The value to test |

Examples

```console
$ '{{ math.IsNum "foo" }} {{ math.IsNum 0xDeadBeef }}'
false true
```

## `Max`

Returns the largest number provided. If any values are floating-point numbers, a `float64` is returned, otherwise an `int64` is returned. The same special-cases as Go's [`math.Max`](https://golang.org/pkg/math/#Max) are followed.


Arguments

| name | description |
|------|-------------|
| `nums...` | _(required)_ One or more numbers to compare |

Examples

```console
$ '{{ math.Max 0 8.0 4.5 "-1.5e-11" }}'
8
```

## `Min`

Returns the smallest number provided. If any values are floating-point numbers, a `float64` is returned, otherwise an `int64` is returned. The same special-cases as Go's [`math.Min`](https://golang.org/pkg/math/#Min) are followed.

Arguments

| name | description |
|------|-------------|
| `nums...` | _(required)_ One or more numbers to compare |

Examples

```console
$ '{{ math.Min 0 8 4.5 "-1.5e-11" }}'
-1.5e-11
```

## `Mul`

## `mul`

Multiply all given operators together.

Arguments

| name | description |
|------|-------------|
| `n...` | _(required)_ The numbers to multiply |

Examples

```console
$ '{{ math.Mul 8 8 2 }}'
128
```


## `pow`

Calculate an exponent - _b<sup>n</sup>_. This wraps Go's [`math.Pow`](https://golang.org/pkg/math/#Pow). If any values are floating-point numbers, a `float64` is returned, otherwise an `int64` is returned.



Arguments

| name | description |
|------|-------------|
| `b` | _(required)_ The base |
| `n` | _(required)_ The exponent |

Examples

```console
$ '{{ math.Pow 10 2 }}'
100
$ '{{ math.Pow 2 32 }}'
4294967296
$ '{{ math.Pow 1.5 2 }}'
2.2
```


## `rem`

Return the remainder from an integer division operation.


Arguments

| name | description |
|------|-------------|
| `a` | _(required)_ The divisor |
| `b` | _(required)_ The dividend |

Examples

```console
$ '{{ math.Rem 5 3 }}'
2
$ '{{ math.Rem -5 3 }}'
-2
```

## `Round`

Returns the nearest integer, rounding half away from zero.

**Note:** the return value of this function is a `float64` so that the special-cases `NaN` and `Inf` can be returned appropriately.


Arguments

| name | description |
|------|-------------|
| `num` | _(required)_ The input number. Will be converted to a `float64`, or `0` if not convertable |

Examples

```console
$ '{{ range (slice -6.5 5.1 42.9 "3.5" 6.5) }}round {{ printf "%#v" . }} = {{ math.Round . }}{{"\n"}}{{ end }}'
round -6.5 = -7
round 5.1 = 5
round 42.9 = 43
round "3.5" = 4
round 6.5 = 7
```

## `seq`

Return a sequence from `start` to `end`, in steps of `step`. Can handle counting down as well as up, including with negative numbers.
Note that the sequence _may_ not end at `end`, if `end` is not divisible by `step`.


Arguments

| name | description |
|------|-------------|
| `start` | _(optional)_ The first number in the sequence (defaults to `1`) |
| `end` | _(required)_ The last number in the sequence |
| `step` | _(optional)_ The amount to increment between each number (defaults to `1`) |

Examples

```console
$ '{{ range (math.Seq 5) }}{{.}} {{end}}'
1 2 3 4 5
```
```console
$ '{{ conv.Join (math.Seq 10 -3 2) ", " }}'
10, 8, 6, 4, 2, 0, -2
```

## `sub`

Subtract the second from the first of the given operators.  When one of the inputs is a floating-point number, the result will be a `float64`, otherwise it will be an `int64`.

Arguments

| name | description |
|------|-------------|
| `a` | _(required)_ The minuend (the number to subtract from) |
| `b` | _(required)_ The subtrahend (the number being subtracted) |

Examples

```console
$ '{{ math.Sub 3 1 }}'
2
```

# `Path`
## `Base`

Returns the last element of path. Trailing slashes are removed before extracting the last element. If the path is empty, Base returns `.`. If the path consists entirely of slashes, Base returns `/`.

A wrapper for Go's [`path.Base`](https://golang.org/pkg/path/#Base) function.

Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ path.Base "/tmp/foo" }}'
foo
```

## `Clean`

Clean returns the shortest path name equivalent to path by purely lexical processing.

A wrapper for Go's [`path.Clean`](https://golang.org/pkg/path/#Clean) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ path.Clean "/tmp//foo/../" }}'
/tmp
```

## `Dir`

Returns all but the last element of path, typically the path's directory.

A wrapper for Go's [`path.Dir`](https://golang.org/pkg/path/#Dir) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ path.Dir "/tmp/foo" }}'
/tmp
```

## `Ext`

Returns the file name extension used by path.

A wrapper for Go's [`path.Ext`](https://golang.org/pkg/path/#Ext) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ path.Ext "/tmp/foo.csv" }}'
.csv
```

## `IsAbs`

Reports whether the path is absolute.

A wrapper for Go's [`path.IsAbs`](https://golang.org/pkg/path/#IsAbs) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ 'the path is {{ if (path.IsAbs "/tmp/foo.csv") }}absolute{{else}}relative{{end}}'
the path is absolute
$ 'the path is {{ if (path.IsAbs "../foo.csv") }}absolute{{else}}relative{{end}}'
the path is relative
```

## `Join`

Joins any number of path elements into a single path, adding a separating slash if necessary.

A wrapper for Go's [`path.Join`](https://golang.org/pkg/path/#Join) function.


Arguments

| name | description |
|------|-------------|
| `elem...` | _(required)_ The path elements to join (0 or more) |

Examples

```console
$ '{{ path.Join "/tmp" "foo" "bar" }}'
/tmp/foo/bar
```

## `Match`

Reports whether name matches the shell file name pattern.

A wrapper for Go's [`path.Match`](https://golang.org/pkg/path/#Match) function.


Arguments

| name | description |
|------|-------------|
| `pattern` | _(required)_ The pattern to match on |
| `path` | _(required)_ The path to match |

Examples

```console
$ '{{ path.Match "*.csv" "foo.csv" }}'
true
```

## `Split`

Splits path immediately following the final slash, separating it into a directory and file name component.

The function returns an array with two values, the first being the directory, and the second the file.

A wrapper for Go's [`path.Split`](https://golang.org/pkg/path/#Split) function.


Arguments

| name | description |
|------|-------------|
| `path` | _(required)_ The input path |

Examples

```console
$ '{{ $p := path.Split "/tmp/foo" }}{{ $dir := index $p 0 }}{{ $file := index $p 1 }}dir is {{$dir}}, file is {{$file}}' dir is /tmp/, file is foo
```

# `Random`
## `ASCII`

Generates a random string of a desired length, containing the set of printable characters from the 7-bit [ASCII](https://en.wikipedia.org/wiki/ASCII) set. This includes _space_ (' '), but no other whitespace characters.


Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |

Examples

```console
$ '{{ random.ASCII 8 }}'
_woJ%D&K
```

## `Alpha`

Generates a random alphabetical (`A-Z`, `a-z`) string of a desired length.


Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |

Examples

```console
$ '{{ random.Alpha 42 }}'
oAqHKxHiytYicMxTMGHnUnAfltPVZDhFkVkgDvatJK
```

## `AlphaNum`

Generates a random alphanumeric (`0-9`, `A-Z`, `a-z`) string of a desired length.


Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |

Examples

```console
$ '{{ random.AlphaNum 16 }}'
4olRl9mRmVp1nqSm
```

## `String`

Generates a random string of a desired length.

By default, the possible characters are those represented by the regular expression `[a-zA-Z0-9_.-]` (alphanumeric, plus `_`, `.`, and `-`).

A different set of characters can be specified with a regular expression, or by giving a range of possible characters by specifying the lower and upper bounds. Lower/upper bounds can be specified as characters (e.g. `"q"`, or escape sequences such as `"\U0001f0AF"`), or numeric Unicode code-points (e.g. `48` or `0x30` for the character `0`).

When given a range of Unicode code-points, `random.String` will discard non-printable characters from the selection. This may result in a much smaller set of possible characters than intended, so check the [Unicode character code charts](http://www.unicode.org/charts/) to verify the correct code-points.


Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the length of the string to produce (number of characters) |
| `regex` | _(optional)_ the regular expression that each character must match (defaults to `[a-zA-Z0-9_.-]`) |
| `lower` | _(optional)_ lower bound for a range of characters (number or single character) |
| `upper` | _(optional)_ upper bound for a range of characters (number or single character) |

Examples

```console
$ '{{ random.String 8 }}'
FODZ01u_
```
```console
$ '{{ random.String 16 `[[:xdigit:]]` }}'
B9e0527C3e45E1f3
```
```console
$ '{{ random.String 20 `[\p{Canadian_Aboriginal}]` }}'
ᗄᖖᣡᕔᕫᗝᖴᒙᗌᘔᓰᖫᗵᐕᗵᙔᗠᓅᕎᔹ
```
```console
$ '{{ random.String 8 "c" "m" }}'
ffmidgjc
```
```console
$ 'You rolled... {{ random.String 3 "⚀" "⚅" }}'
You rolled... ⚅⚂⚁
```
```console
$ 'Poker time! {{ random.String 5 "\U0001f0a1" "\U0001f0de" }}'
Poker time! 🂼🂺🂳🃅🂪
```

## `Item`

Pick an element at a random from a given slice or array.


Arguments

| name | description |
|------|-------------|
| `items` | _(required)_ the input array |

Examples

```console
$ '{{ random.Item (seq 0 5) }}'
4
```
```console
$ export SLICE='["red", "green", "blue"]'
$ '{{ getenv "SLICE" | jsonArray | random.Item }}'
blue
```

## `Number`

Pick a random integer. By default, a number between `0` and `100` (inclusive) is chosen, but this range can be overridden.

Note that the difference between `min` and `max` can not be larger than a 63-bit integer (i.e. the unsigned portion of a 64-bit signed integer). The result is given as an `int64`.

Arguments

| name | description |
|------|-------------|
| `min` | _(optional)_ The minimum value, defaults to `0`. Must be less than `max`. |
| `max` | _(optional)_ The maximum value, defaults to `100` (if no args provided) |

Examples

```console
$ '{{ random.Number }}'
55
```
```console
$ '{{ random.Number -10 10 }}'
-3
```
```console
$ '{{ random.Number 5 }}'
2
```

## `Float`

Pick a random decimal floating-point number. By default, a number between `0.0` and `1.0` (_exclusive_, i.e. `[0.0,1.0)`) is chosen, but this range can be overridden.

The result is given as a `float64`.

Arguments

| name | description |
|------|-------------|
| `min` | _(optional)_ The minimum value, defaults to `0.0`. Must be less than `max`. |
| `max` | _(optional)_ The maximum value, defaults to `1.0` (if no args provided). |

Examples

```console
$ '{{ random.Float }}'
0.2029946480303966
```
```console
$ '{{ random.Float 100 }}'  
71.28595374161743
```
```console
$ '{{ random.Float -100 200 }}'
105.59119437834909
```

# `regexp`
## `Find`

Returns a string holding the text of the leftmost match in `input` of the regular expression `expression`.

This function provides the same behaviour as Go's
[`regexp.FindString`](https://golang.org/pkg/regexp/#Regexp.FindString) function.


Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `input` | _(required)_ The input to search |

Examples

```console
$ '{{ regexp.Find "[a-z]{3}" "foobar"}}'
foo
```
```console
$ 'no {{ "will not match" | regexp.Find "[0-9]" }}numbers'
no numbers
```

## `FindAll`

Returns a list of all successive matches of the regular expression.

This can be called with 2 or 3 arguments. When called with 2 arguments, the `n` argument (number of matches) will be set to `-1`, causing all matches to be returned.

This function provides the same behaviour as Go's
[`regexp.FindAllString`](https://golang.org/pkg/regexp/#Regexp.FindAllString) function.


Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `n` | _(optional)_ The number of matches to return |
| `input` | _(required)_ The input to search |

Examples

```console
$ '{{ regexp.FindAll "[a-z]{3}" "foobar" | toJSON}}'
["foo", "bar"]
```
```console
$ '{{ "foo bar baz qux" | regexp.FindAll "[a-z]{3}" 3 | toJSON}}'
["foo", "bar", "baz"]
```

## `Match`

Returns `true` if a given regular expression matches a given input.

This returns a boolean which can be used in an `if` condition, for example.

Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `input` | _(required)_ The input to test |

Examples

```console
$ '{{ if (.Env.USER | regexp.Match `^h`) }}username ({{.Env.USER}}) starts with h!{{end}}'
username (hairyhenderson) starts with h!
```

## `QuoteMeta`

Escapes all regular expression metacharacters in the input. The returned string is a regular expression matching the literal text.

This function provides the same behaviour as Go's
[`regexp.QuoteMeta`](https://golang.org/pkg/regexp/#Regexp.QuoteMeta) function.


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ The input to escape |

Examples

```console
$ '{{ `{hello}` | regexp.QuoteMeta }}'
\{hello\}
```

## `Replace`

Replaces matches of a regular expression with the replacement string.

The replacement is substituted after expanding variables beginning with `$`.

This function provides the same behaviour as Go's [`regexp.ReplaceAllString`](https://golang.org/pkg/regexp/#Regexp.ReplaceAllString) function.


Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression string |
| `replacement` | _(required)_ The replacement string |
| `input` | _(required)_ The input string to operate on |

Examples

```console
$ '{{ regexp.Replace "(foo)bar" "$1" "foobar"}}'
foo
```
```console
$ '{{ regexp.Replace "(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)" "${last}, ${first}" "Alan Turing"}}'
Turing, Alan
```

## `ReplaceLiteral`

Replaces matches of a regular expression with the replacement string.

The replacement is substituted directly, without expanding variables
beginning with `$`.

This function provides the same behaviour as Go's [`regexp.ReplaceAllLiteralString`](https://golang.org/pkg/regexp/#Regexp.ReplaceAllLiteralString) function.


Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression string |
| `replacement` | _(required)_ The replacement string |
| `input` | _(required)_ The input string to operate on |

Examples

```console
$ '{{ regexp.ReplaceLiteral "(foo)bar" "$1" "foobar"}}'
$1
```
```console
$ '{{ `foo.bar,baz` | regexp.ReplaceLiteral `\W` `$` }}'
foo$bar$baz
```

# `Strings`
## `Split`

Splits `input` into sub-strings, separated by the expression.

This can be called with 2 or 3 arguments. When called with 2 arguments, the `n` argument (number of matches) will be set to `-1`, causing all sub-strings to be returned.

This is equivalent to [`strings.SplitN`](../strings/#strings-splitn), except that regular expressions are supported.

This function provides the same behaviour as Go's [`regexp.Split`](https://golang.org/pkg/regexp/#Regexp.Split) function.


Arguments

| name | description |
|------|-------------|
| `expression` | _(required)_ The regular expression |
| `n` | _(optional)_ The number of matches to return |
| `input` | _(required)_ The input to search |

Examples

```console
$ '{{ regexp.Split `[\s,.]` "foo bar,baz.qux" | toJSON}}'
["foo","bar","baz","qux"]
```
```console
$ '{{ "foo bar.baz,qux" | regexp.Split `[\s,.]` 3 | toJSON}}'
["foo","bar","baz"]
```


## `Abbrev`

Abbreviates a string using `...` (ellipses). Takes an optional offset from the beginning of the string, and a maximum final width (including added ellipses).

_Also see [`strings.Trunc`](#strings-trunc)._


Arguments

| name | description |
|------|-------------|
| `offset` | _(optional)_ offset from the start of the string. Must be `4` or greater for ellipses to be added. Defaults to `0` |
| `width` | _(required)_ the desired maximum final width of the string, including ellipses |
| `input` | _(required)_ the input string to abbreviate |

Examples

```console
$ '{{ "foobarbazquxquux" | strings.Abbrev 9 }}'
foobar...
$ '{{ "foobarbazquxquux" | strings.Abbrev 6 9 }}'
...baz...
```

## `Contains`

Reports whether a substring is contained within a string.

Arguments

| name | description |
|------|-------------|
| `substr` | _(required)_ the substring to search for |
| `input` | _(required)_ the input to search |

Examples

_`input.tmpl`:_
```
FOO=foo 
{{ if (.Env.FOO | strings.Contains "f") }}yes{{else}}no{{end}}
```

```console
yes
```

## `HasPrefix`

Tests whether a string begins with a certain prefix.


Arguments

| name | description |
|------|-------------|
| `prefix` | _(required)_ the prefix to search for |
| `input` | _(required)_ the input to search |

Examples

```console
$ URL=http://example.com 
'{{if .Env.URL | strings.HasPrefix "https"}}foo{{else}}bar{{end}}'
bar
$ URL=https://example.com
'{{if .Env.URL | strings.HasPrefix "https"}}foo{{else}}bar{{end}}'
foo
```

## `HasSuffix`

Tests whether a string ends with a certain suffix.


Arguments

| name | description |
|------|-------------|
| `suffix` | _(required)_ the suffix to search for |
| `input` | _(required)_ the input to search |

Examples

_`input.tmpl`:_
```
URL=http://example.com
{{.Env.URL}}{{if not (.Env.URL | strings.HasSuffix ":80")}}:80{{end}}
```

```console
http://example.com:80
```

## `Indent`

## `indent`

Indents a string. If the input string has multiple lines, each line will be indented.


Arguments

| name | description |
|------|-------------|
| `width` | _(optional)_ number of times to repeat the `indent` string. Default: `1` |
| `indent` | _(optional)_ the string to indent with. Default: `" "` |
| `input` | _(required)_ the string to indent |

Examples

This function can be especially useful when adding YAML snippets into other YAML documents, where indentation is important:

_`input.tmpl`:_
```
foo:
{{ `{"bar": {"baz": 2}}` | json | toYAML | strings.Indent "  " }}
{{- `{"qux": true}` | json | toYAML | strings.Indent 2 }}
  quux:
{{ `{"quuz": 42}` | json | toYAML | strings.Indent 2 "  " -}}
```

```console
foo:
  bar:
    baz: 2
  qux: true

  quux:
    quuz: 42
```

## `Sort` _(deprecated)_
**Deprecation Notice:** Use [`coll.Sort`](../coll/#coll-sort) instead

Returns an alphanumerically-sorted copy of a given string list.


Arguments

| name | description |
|------|-------------|
| `list` | _(required)_ The list to sort |

Examples

```console
$ '{{ (slice "foo" "bar" "baz") | strings.Sort }}'
[bar baz foo]
```

## `Split`

Creates a slice by splitting a string on a given delimiter.

Arguments

| name | description |
|------|-------------|
| `separator` | _(required)_ the string sequence to split |
| `input` | _(required)_ the input string |

Examples

Use on its own to produce an array:
```console
$ '{{ "Bart,Lisa,Maggie" | strings.Split "," }}'
[Bart Lisa Maggie]
```

Use in combination with `range` to iterate over all items:
```console
$ '{{range ("Bart,Lisa,Maggie" | strings.Split ",") }}Hello, {{.}}
{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

Use in combination with `index` function to pick a specific value from the resulting array
```console
$ '{{index ("Bart,Lisa,Maggie" | strings.Split ",") 0 }}'
Bart
```


## `SplitN`

Creates a slice by splitting a string on a given delimiter. The count determines the number of substrings to return.


Arguments

| name | description |
|------|-------------|
| `separator` | _(required)_ the string sequence to split |
| `count` | _(required)_ the maximum number of substrings to return |
| `input` | _(required)_ the input string |

Examples

```console
$ '{{ range ("foo:bar:baz" | strings.SplitN ":" 2) }}{{.}}
{{end}}'
foo
bar:baz
```


## `quote`

Surrounds an input string with double-quote characters (`"`). If the input is not a string, converts first.

`"` characters in the input are first escaped with a `\` character.

This is a convenience function which is equivalent to:

```
{{ print "%q" "input string" }}
```


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input to quote |

Examples

```console
$ '{{ "in" | quote }}'
"in"
$ '{{ strings.Quote 500 }}'
"500"
```

## `Repeat`

Returns a new string consisting of `count` copies of the input string.

It errors if `count` is negative or if the length of `input` multiplied by `count` overflows.

This wraps Go's [`strings.Repeat`](https://golang.org/pkg/strings/#Repeat).


Arguments

| name | description |
|------|-------------|
| `count` | _(required)_ the number of times to repeat the input |
| `input` | _(required)_ the input to repeat |

Examples

```console
$ '{{ "hello " | strings.Repeat 5 }}'
hello hello hello hello hello
```


## `replaceAll`

Replaces all occurrences of a given string with another.


Arguments

| name | description |
|------|-------------|
| `old` | _(required)_ the text to replace |
| `new` | _(required)_ the new text to replace with |
| `input` | _(required)_ the input to modify |

Examples

```console
$ '{{ strings.ReplaceAll "." "-" "172.21.1.42" }}'
172-21-1-42
$ '{{ "172.21.1.42" | strings.ReplaceAll "." "-" }}'
172-21-1-42
```

## `Slug`

Creates a a "slug" from a given string - supports Unicode correctly. This wraps the [github.com/gosimple/slug](https://github.com/gosimple/slug) package. See [the github.com/gosimple/slug docs](https://godoc.org/github.com/gosimple/slug) for more information.


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input to "slugify" |

Examples

```console
$ '{{ "Hello, world!" | strings.Slug }}'
hello-world
```

## `shellQuote`

Given a string, emits a version of that string that will evaluate to its literal data when expanded by any POSIX-compliant shell.
Given an array or slice, emit a single string which will evaluate to a series of shell words, one per item in that array or slice.

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input to quote |

Examples

```console
$ "{{ slice \"one word\" \"foo='bar baz'\" | shellQuote }}"
'one word' 'foo='"'"'bar baz'"'"''
```
```console
$ "{{ strings.ShellQuote \"it's a banana\" }}"
'it'"'"'s a banana'
```


## `squote`

Surrounds an input string with a single-quote (apostrophe) character (`'`). If the input is not a string, converts first.

`'` characters in the input are first escaped in the YAML-style (by repetition: `''`).

Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input to quote |

Examples

```console
$ '{{ "in" | squote }}'
'in'
```
```console
$ "{{ strings.Squote \"it's a banana\" }}"
'it''s a banana'
```


## `title`

Convert to title-case.


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

Examples

```console
$ '{{strings.Title "hello, world!"}}'
Hello, World!
```


## `toLower`

Convert to lower-case.


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

Examples

```console
$ echo '{{strings.ToLower "HELLO, WORLD!"}}'
hello, world!
```


## `toUpper`

Convert to upper-case.

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

Examples

```console
$ '{{strings.ToUpper "hello, world!"}}'
HELLO, WORLD!
```

## `Trim`

Trims a string by removing the given characters from the beginning and end of
the string.


Arguments

| name | description |
|------|-------------|
| `cutset` | _(required)_ the set of characters to cut |
| `input` | _(required)_ the input |

Examples

```console
$ '{{ "_-foo-_" | strings.Trim "_-" }}
foo
```

## `TrimPrefix`

Returns a string without the provided leading prefix string, if the prefix is present.

This wraps Go's [`strings.TrimPrefix`](https://golang.org/pkg/strings/#TrimPrefix).

Arguments

| name | description |
|------|-------------|
| `prefix` | _(required)_ the prefix to trim |
| `input` | _(required)_ the input |

Examples

```console
$ '{{ "hello, world" | strings.TrimPrefix "hello, " }}'
world
```


## `trimSpace`

Trims a string by removing whitespace from the beginning and end of
the string.

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |

Examples

```console
$ '{{ "  \n\t foo" | strings.TrimSpace }}'
foo
```

## `TrimSuffix`

Returns a string without the provided trailing suffix string, if the suffix is present.

This wraps Go's [`strings.TrimSuffix`](https://golang.org/pkg/strings/#TrimSuffix).



Arguments

| name | description |
|------|-------------|
| `suffix` | _(required)_ the suffix to trim |
| `input` | _(required)_ the input |

Examples

```console
$ '{{ "hello, world" | strings.TrimSuffix "world" }}jello'
hello, jello
```

## `Trunc`

Returns a string truncated to the given length.

_Also see [`strings.Abbrev`](#strings-abbrev)._

Arguments

| name | description |
|------|-------------|
| `length` | _(required)_ the maximum length of the output |
| `input` | _(required)_ the input |

Examples

```console
$ '{{ "hello, world" | strings.Trunc 5 }}'
hello
```

## `CamelCase`

Converts a sentence to CamelCase, i.e. `The quick brown fox` becomes `TheQuickBrownFox`.

All non-alphanumeric characters are stripped, and the beginnings of words are upper-cased. If the input begins with a lower-case letter, the result will also begin with a lower-case letter.

See [CamelCase on Wikipedia](https://en.wikipedia.org/wiki/Camel_case) for more details.


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input |

Examples

```console
$ '{{ "Hello, World!" | strings.CamelCase }}'
HelloWorld
```
```console
$ '{{ "hello jello" | strings.CamelCase }}'
helloJello
```

## `SnakeCase`

Converts a sentence to snake_case, i.e. `The quick brown fox` becomes `The_quick_brown_fox`.

All non-alphanumeric characters are stripped, and spaces are replaced with an underscore (`_`). If the input begins with a lower-case letter, the result will also begin with a lower-case letter.

See [Snake Case on Wikipedia](https://en.wikipedia.org/wiki/Snake_case) for more details.


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input |

Examples

```console
$ '{{ "Hello, World!" | strings.SnakeCase }}'
Hello_world
```
```console
$ '{{ "hello jello" | strings.SnakeCase }}'
hello_jello
```

## `KebabCase`

Converts a sentence to kebab-case, i.e. `The quick brown fox` becomes `The-quick-brown-fox`. All non-alphanumeric characters are stripped, and spaces are replaced with a hyphen (`-`). If the input begins with a lower-case letter, the result will also begin with a lower-case letter.
See [Kebab Case on Wikipedia](https://en.wikipedia.org/wiki/Kebab_case) for more details.


Arguments

| name | description |
|------|-------------|
| `in` | _(required)_ The input |

Examples

```console
$ '{{ "Hello, World!" | strings.KebabCase }}'
Hello-world
```
```console
$ '{{ "hello jello" | strings.KebabCase }}'
hello-jello
```

## `WordWrap`

Inserts new line breaks into the input string so it ends up with lines that are at most `width` characters wide. The line-breaking algorithm is _naïve_ and _greedy_: lines are only broken between words (i.e. on whitespace characters), and no effort is made to "smooth" the line endings. When words that are longer than the desired width are encountered (e.g. long URLs), they are not broken up. Correctness is valued above line length.

The line-break sequence defaults to `\n` (i.e. the LF/Line Feed character), regardless of OS.


Arguments

| name | description |
|------|-------------|
| `width` | _(optional)_ The desired maximum line length (number of characters - defaults to `80`) |
| `lbseq` | _(optional)_ The line-break sequence to use (defaults to `\n`) |
| `in` | _(required)_ The input |

Examples

```console
$ '{{ "Hello, World!" | strings.WordWrap 7 }}'
Hello,
World!
```
```console
$ '{{ strings.WordWrap 20 "\\\n" "a string with a long url http://example.com/a/very/long/url which should not be broken" }}'
a string with a long
url
http://example.com/a/very/long/url
which should not be
broken
```

## `RuneCount`

Return the number of _runes_ (Unicode code-points) contained within the input. This is similar to the built-in `len` function, but `len` counts the length in _bytes_. The length of an input containing multi-byte code-points should therefore be measured with `strings.RuneCount`.

Inputs will first be converted to strings, and multiple inputs are concatenated.

This wraps Go's [`utf8.RuneCountInString`](https://golang.org/pkg/unicode/utf8/#RuneCountInString) function.

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input(s) to measure |

Examples

```console
$ '{{ range (slice "\u03a9" "\u0030" "\u1430") }}{{ printf "%s is %d bytes and %d runes\n" . (len .) (strings.RuneCount .) }}{{ end }}'
Ω is 2 bytes and 1 runes
0 is 1 bytes and 1 runes
ᐰ is 3 bytes and 1 runes
```

## `contains`

**See [`strings.Contains`](#strings-contains) for a pipeline-compatible version**

Contains reports whether the second string is contained within the first. Equivalent to [strings.Contains](https://golang.org/pkg/strings#Contains)


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the string to search |
| `substring` | _(required)_ the string to search for |

Examples

_`input.tmpl`:_
```
FOO=foo
{{if contains .Env.FOO "f"}}yes{{else}}no{{end}}
```

```console
yes
```

## `hasPrefix`

**See [`strings.HasPrefix`](#strings-hasprefix) for a pipeline-compatible version**

Tests whether the string begins with a certain substring. Equivalent to [strings.HasPrefix](https://golang.org/pkg/strings#HasPrefix)



Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the string to search |
| `prefix` | _(required)_ the prefix to search for |

Examples

_`input.tmpl`:_
```
URL=http://example.com
{{if hasPrefix .Env.URL "https"}}foo{{else}}bar{{end}}
```

```console
bar
```

## `hasSuffix`

**See [`strings.HasSuffix`](#strings-hassuffix) for a pipeline-compatible version**

Tests whether the string ends with a certain substring. Equivalent to [strings.HasSuffix](https://golang.org/pkg/strings#HasSuffix)


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input to search |
| `suffix` | _(required)_ the suffix to search for |

Examples

_`input.tmpl`:_
```
URL=http://example.com
{{.Env.URL}}{{if not (hasSuffix .Env.URL ":80")}}:80{{end}}
```

```console
http://example.com:80
```

## `split`

**See [`strings.Split`](#strings-split) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. Equivalent to [strings.Split](https://golang.org/pkg/strings#Split)

Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input string |
| `separator` | _(required)_ the string sequence to split |

Examples

```console
$ '{{range split "Bart,Lisa,Maggie" ","}}Hello, {{.}}
{{end}}'
Hello, Bart
Hello, Lisa
Hello, Maggie
```

## `splitN`

**See [`strings.SplitN`](#strings-splitn) for a pipeline-compatible version**

Creates a slice by splitting a string on a given delimiter. The count determines the number of substrings to return. Equivalent to [strings.SplitN](https://golang.org/pkg/strings#SplitN)


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input string |
| `separator` | _(required)_ the string sequence to split |
| `count` | _(required)_ the maximum number of substrings to return |

Examples

```console
$ '{{ range splitN "foo:bar:baz" ":" 2 }}{{.}}
{{end}}'
foo
bar:baz
```

## `trim`

**See [`strings.Trim`](#strings-trim) for a pipeline-compatible version**

Trims a string by removing the given characters from the beginning and end of the string. Equivalent to [strings.Trim](https://golang.org/pkg/strings/#Trim)


Arguments

| name | description |
|------|-------------|
| `input` | _(required)_ the input |
| `cutset` | _(required)_ the set of characters to cut |

Examples

_`input.tmpl`:_
```
FOO="  world "
Hello, {{trim .Env.FOO " "}}!
```

```console
Hello, world!
```


# `test`

## `assert`

Asserts that the given expression or value is `true`. If it is not, causes template generation to fail immediately with an optional message.

Arguments

| name | description |
|------|-------------|
| `message` | _(optional)_ The optional message to provide in the case of failure |
| `value` | _(required)_ The value to test |

Examples

```console
$ '{{ assert (eq "foo" "bar") }}'
template: <arg>:1:3: executing "<arg>" at <assert (eq "foo" "ba...>: error calling assert: assertion failed
$ '{{ assert "something horrible happened" false }}'
template: <arg>:1:3: executing "<arg>" at <assert "something ho...>: error calling assert: assertion failed: something horrible happened
```


## `fail`

Cause template generation to fail immediately, with an optional message.


Arguments

| name | description |
|------|-------------|
| `message` | _(optional)_ The optional message to provide |

Examples

```console
$ '{{ fail }}'
template: <arg>:1:3: executing "<arg>" at <fail>: error calling fail: template generation failed
$ '{{ test.Fail "something is wrong!" }}'
template: <arg>:1:7: executing "<arg>" at <test.Fail>: error calling Fail: template generation failed: something is wrong!
```

## `isKind`

Report whether the argument is of the given Kind. Can be used to render different templates depending on the kind of data.

See [the Go `reflect` source code](https://github.com/golang/go/blob/36fcde1676a0d3863cb5f295eed6938cd782fcbb/src/reflect/type.go#L595..L622) for the complete list, but these are some common values:

- `string`
- `bool`
- `int`, `int64`, `uint64`
- `float64`
- `slice`
- `map`
- `invalid` (a catch-all, usually just `nil` values)

In addition, the special kind `number` is accepted by this function, to represent _any_ numeric kind (whether `float32`, `uint8`, or whatever). This is useful when the specific numeric type is unknown.

See also [`test.Kind`](test-kind).

Arguments

| name | description |
|------|-------------|
| `kind` | _(required)_ the kind to compare with (see desription for possible values) |
| `value` | _(required)_ the value to check |

Examples

```console
$ '{{ $data := "hello world" }}
{{- if isKind "string" $data }}{{ $data }} is a string{{ end }}'
hello world is a string
```
```console
$ '{{ $object := dict "key1" true "key2" "foobar" }}
{{- if test.IsKind "map" $object }}
Got a map:
{{ range $key, $value := $object -}}
  - "{{ $key }}": {{ $value }}
{{ end }}
{{ else if test.IsKind "number" $object }}
Got a number: {{ $object }}
{{ end }}'

Got a map:
- "key1": true
- "key2": foobar
```

## `kind`

Report the _kind_ of the given argument. This differs from the _type_ of the argument in specificity; for example, while a slice of strings may have a type of `[]string`, the _kind_ of that slice will simply be `slice`.
If you need to know the precise type of a value, use `printf "%T" $value`.

See also [`test.IsKind`](test-iskind).


Arguments

| name | description |
|------|-------------|
| `value` | _(required)_ the value to check |

Examples

```console
$ '{{ kind "hello world" }}'
string
```
```console
$ '{{ dict "key1" true "key2" "foobar" | test.Kind }}'
map
```


## `required`

Passes through the given value, if it's non-empty, and non-`nil`. Otherwise, exits and prints a given error message so the user can adjust as necessary. This is particularly useful for cases where templates require user-provided data (such as datasources or environment variables), and rendering can not continue correctly.

This was inspired by [Helm's `required` function](https://github.com/kubernetes/helm/blob/master/docs/charts_tips_and_tricks.md#know-your-template-functions), but has slightly different behaviour. 

Arguments

| name | description |
|------|-------------|
| `message` | _(optional)_ The optional message to provide when the required value is not provided |
| `value` | _(required)_ The required value |

Examples

```console
$ FOO=foobar 
'{{ getenv "FOO" | required "Missing FOO environment variable!" }}'
foobar
$ FOO= 
'{{ getenv "FOO" | required "Missing FOO environment variable!" }}'
error: Missing FOO environment variable!
```

## `ternary`

Returns one of two values depending on whether the third is true. Note that the third value does not have to be a boolean - it is converted first by the [`conv.ToBool`](../conv/#conv-tobool) function (values like `true`, `1`, `"true"`, `"Yes"`, etc... are considered true).

This is effectively a short-form of the following template:

```
{{ if conv.ToBool $condition }}{{ $truevalue }}{{ else }}{{ $falsevalue }}{{ end }}
```

Keep in mind that using an explicit `if`/`else` block is often easier to understand than ternary expressions!


Arguments

| name | description |
|------|-------------|
| `truevalue` | _(required)_ the value to return if `condition` is true |
| `falsevalue` | _(required)_ the value to return if `condition` is false |
| `condition` | _(required)_ the value to evaluate for truthiness |

Examples

```console
$ '{{ ternary "FOO" "BAR" false }}'
BAR
$ '{{ ternary "FOO" "BAR" "yes" }}'
FOO
```

# `Time`
## `Now`

Returns the current local time, as a `time.Time`. This wraps [`time.Now`](https://golang.org/pkg/time/#Now).

Usually, further functions are called using the value returned by `Now`.


Examples

Usage with [`UTC`](https://golang.org/pkg/time/#Time.UTC) and [`Format`](https://golang.org/pkg/time/#Time.Format):
```console
$ '{{ (time.Now).UTC.Format "Day 2 of month 1 in year 2006 (timezone MST)" }}'
Day 14 of month 10 in year 2017 (timezone UTC)
```
Usage with [`AddDate`](https://golang.org/pkg/time/#Time.AddDate):
```console
$ date
Sat Oct 14 09:57:02 EDT 2017
$ '{{ ((time.Now).AddDate 0 1 0).Format "Mon Jan 2 15:04:05 MST 2006" }}'
Tue Nov 14 09:57:02 EST 2017
```

_(notice how the TZ adjusted for daylight savings!)_
Usage with [`IsDST`](https://golang.org/pkg/time/#Time.IsDST):
```console
$ '{{ $t := time.Now }}At the tone, the time will be {{ ($t.Round (time.Minute 1)).Add (time.Minute 1) }}.
  It is{{ if not $t.IsDST }} not{{ end }} daylight savings time.
  ... ... BEEP'
At the tone, the time will be 2022-02-10 09:01:00 -0500 EST.
It is not daylight savings time.
... ... BEEP
```

## `Parse`

Parses a timestamp defined by the given layout. This wraps [`time.Parse`](https://golang.org/pkg/time/#Parse).

A number of pre-defined layouts are provided as constants, defined
[here](https://golang.org/pkg/time/#pkg-constants).

Just like [`time.Now`](#time-now), this is usually used in conjunction with other functions.

_Note: In the absence of a time zone indicator, `time.Parse` returns a time in UTC._


Arguments

| name | description |
|------|-------------|
| `layout` | _(required)_ The layout string to parse with |
| `timestamp` | _(required)_ The timestamp to parse |

Examples

Usage with [`Format`](https://golang.org/pkg/time/#Time.Format):
```console
$ '{{ (time.Parse "2006-01-02" "1993-10-23").Format "Monday January 2, 2006 MST" }}'
Saturday October 23, 1993 UTC
```

## `ParseDuration`

Parses a duration string. This wraps [`time.ParseDuration`](https://golang.org/pkg/time/#ParseDuration).

A duration string is a possibly signed sequence of decimal numbers, each with
optional fraction and a unit suffix, such as `300ms`, `-1.5h` or `2h45m`. Valid
time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.


Arguments

| name | description |
|------|-------------|
| `duration` | _(required)_ The duration string to parse |

Examples

```console
$ '{{ (time.Now).Format time.Kitchen }}
{{ ((time.Now).Add (time.ParseDuration "2h30m")).Format time.Kitchen }}'
12:43AM
3:13AM
```

## `ParseLocal`

Same as [`time.Parse`](#time-parse), except that in the absence of a time zone indicator, the timestamp wil be parsed in the local timezone.



Arguments

| name | description |
|------|-------------|
| `layout` | _(required)_ The layout string to parse with |
| `timestamp` | _(required)_ The timestamp to parse |

Examples

Usage with [`Format`](https://golang.org/pkg/time/#Time.Format):
```console
$ '{{ (time.ParseLocal time.Kitchen "6:00AM").Format "15:04 MST" }}'
06:00 EST
```

## `ParseInLocation`

Same as [`time.Parse`](#time-parse), except that the time is parsed in the given location's time zone.

This wraps [`time.ParseInLocation`](https://golang.org/pkg/time/#ParseInLocation).


Arguments

| name | description |
|------|-------------|
| `layout` | _(required)_ The layout string to parse with |
| `location` | _(required)_ The location to parse in |
| `timestamp` | _(required)_ The timestamp to parse |

Examples

Usage with [`Format`](https://golang.org/pkg/time/#Time.Format):
```console
$ '{{ (time.ParseInLocation time.Kitchen "Africa/Luanda" "6:00AM").Format "15:04 MST" }}'
06:00 LMT
```

## `Since`

Returns the time elapsed since a given time. This wraps [`time.Since`](https://golang.org/pkg/time/#Since).

It is shorthand for `time.Now.Sub t`.

Arguments

| name | description |
|------|-------------|
| `t` | _(required)_ the `Time` to calculate since |

Examples

```console
$ '{{ $t := time.Parse time.RFC3339 "1970-01-01T00:00:00Z" }}time since the epoch:{{ time.Since $t }}'

time since the epoch:423365h0m24.353828924s
```

## `Unix`

Returns the local `Time` corresponding to the given Unix time, in seconds since January 1, 1970 UTC. Note that fractional seconds can be used to denote milliseconds, but must be specified as a string, not a floating point number.


Arguments

| name | description |
|------|-------------|
| `time` | _(required)_ the time to parse |

Examples

_with whole seconds:_
```console
$ '{{ (time.Unix 42).UTC.Format time.Stamp}}'
Jan  1, 00:00:42
```

_with fractional seconds:_
```console
$ '{{ (time.Unix "123456.789").UTC.Format time.StampMilli}}'
Jan  2 10:17:36.789
```

## `Until`

Returns the duration until a given time. This wraps [`time.Until`](https://golang.org/pkg/time/#Until).

It is shorthand for `$t.Sub time.Now`.

Arguments

| name | description |
|------|-------------|
| `t` | _(required)_ the `Time` to calculate until |

Examples

```console
$ '{{ $t := time.Parse time.RFC3339 "2020-01-01T00:00:00Z" }}only {{ time.Until $t }} to go...'
only 14922h56m46.578625891s to go...
```

Or, less precise:
```console
$ '{{ $t := time.Parse time.RFC3339 "2020-01-01T00:00:00Z" }}only {{ (time.Until $t).Round (time.Hour 1) }} to go...'
only 14923h0m0s to go...
```

## `ZoneName`

Return the local system's time zone's name.

Examples

```console
$ '{{time.ZoneName}}'
EDT
```

## `ZoneOffset`

Return the local system's time zone offset, in seconds east of UTC.

Examples

```console
$ '{{time.ZoneOffset}}'
-14400
```

# `UUID`
## `V1`

Create a version 1 UUID (based on the current MAC address and the current date/time).

Use [`uuid.V4`](#uuid-v4) instead in most cases.


Examples

```console
$ '{{ uuid.V1 }}'
4d757e54-446d-11e9-a8fa-72000877c7b0
```

## `V4`

Create a version 4 UUID (randomly generated).

This function consumes entropy.

Examples

```console
$ '{{ uuid.V4 }}'
40b3c2d2-e491-4b19-94cd-461e6fa35a60
```

## `Nil`

Returns the _nil_ UUID, that is, `00000000-0000-0000-0000-000000000000`,
mostly for testing scenarios.



Examples

```console
$ '{{ uuid.Nil }}'
00000000-0000-0000-0000-000000000000
```

## `IsValid`

Checks that the given UUID is in the correct format. It does not validate whether the version or variant are correct.


Arguments

| name | description |
|------|-------------|
| `uuid` | _(required)_ The uuid to check |

Examples

```console
$ '{{ if uuid.IsValid "totally invalid" }}valid{{ else }}invalid{{ end }}'
invalid
```
```console
$ '{{ uuid.IsValid "urn:uuid:12345678-90ab-cdef-fedc-ba9876543210" }}'
true
```

## `Parse`

Parse a UUID for further manipulation or inspection.

This function returns a `UUID` struct, as defined in the [github.com/google/uuid](https://godoc.org/github.com/google/uuid#UUID) package. See the docs for examples of functions or fields you can call.

Both the standard UUID forms of `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` and `urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` are decoded as well as the Microsoft encoding `{xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}` and the raw hex encoding (`xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`).

Arguments

| name | description |
|------|-------------|
| `uuid` | _(required)_ The uuid to parse |

Examples

```console
$ '{{ $u := uuid.Parse uuid.V4 }}{{ $u.Version }}, {{ $u.Variant}}'
VERSION_4, RFC4122
```
```console
$ '{{ (uuid.Parse "000001f5-4470-21e9-9b00-72000877c7b0").Domain }}'
Person
```
