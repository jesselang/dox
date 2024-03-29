<!-- dox: 1639121104 -->
# Meet dox

dox publishes markdown to Confluence as HTML. For those who prefer to keep
documentation in git, but need to publish to Confluence because... reasons.

## Installing

```sh
go install github.com/jesselang/dox
```

## Use

```sh
# tell dox where to publish
cat - <<EOF > .dox.yaml
uri: https://confluence.yourcompany.com
space: DEMO
title: "title of generated root page"
browse_url_base: https://github.com/jesselang/dox/blob/master
EOF

# set DOX_USERNAME and DOX_PASSWORD to appropriate values
 export DOX_USERNAME=...
 export DOX_PASSWORD=...

# publish
dox [-v]
```

dox will publish **all markdown files** as children of a root page. Each
markdown file will be modified with a dox header. Be sure to commit `.dox.yaml`
and the modified markdown in your source code management.

## dox Header

All markdown files should have a *dox header*. The dox header is a single line
comment at the top of source files that contains a comma separated list of
items. The items include a page ID and optional *source directives*.

All **published** pages will have the page ID included in the dox header.

```
<!-- dox: 1234567890 -->
```

## Source Directives

Source directives allow the user to control how files are published. They are
defined as items in the comma separated list in the dox header.

```
<!-- dox: 1234567890, <directive>, <directive> -->
```

### Ignore Directive

If a file should not be published, use the ignore directive. The ignore
directive should be the only item in the dox header.

```
<!-- dox: ignore -->
```

### Omit Notice Directive

By default, dox adds a notice at the top of each published page stating that
modifications to the page should be done to the source file. This notice can be
omitted.

```
<!-- dox: 1234567890, omit-notice -->
```

## Relative Linking

Websites like github allow markdown files to relatively link to other files in
the repository. dox will try to replicate this functionality by changing
relative links it finds in pages to be published.

If the relative link points to a dox source file, it will change that link to
the URL for that published page (currently limited to Confluence). Otherwise,
dox will change the link to the file's browse URL using `browse_url_base`.

`browse_url_base` supports formatting using the `%s` verb:

```
# dox appends '/%s'
browse_url_base: https://github.com/jesselang/dox/blob/master

browse_url_base: https://othersourcesite.com/repo-base/browse/%s?format=raw
```

<!-- ## Not supported -->

## Roadmap

- Improve [go-confluence][go-confluence] error handling
- Media files
- `dox update` updates already published files without source modifications
- `dox init` to create initial config
- `dox add` to initially publish new files
- `dox ignore` to ignore files
- Consider managing a manifest file in repo using UUIDs instead of
confluence page ID

## Developing

dox requires Go 1.11 modules or vgo. Clone the repo outside of your GOPATH,
and use `go build`.


[go-confluence]: https://github.com/jesselang/go-confluence
