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
EOF

# set DOX_USERNAME and DOX_PASSWORD to appropriate values
 export DOX_USERNAME=...
 export DOX_PASSWORD=...

# publish
dox [-v]
```

dox will publish **all markdown files** as children of a generated root page.
Each markdown file will be modified with a dox header, and `.dox.yaml` will be
updated to include the page_id of the generated root page. Be sure to commit
`.dox.yaml` and these changes to markdown in your source code management.

## Not supported

- Relative links
- Media files

## Roadmap

- Improve [go-confluence][go-confluence] error handling
- Relative links
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
