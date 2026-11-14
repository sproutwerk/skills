# README audit table

`audits` rewrites the audit table in the repository `README.md` from the
results [skills.sh](https://skills.sh/sproutwerk/skills) shows for each skill.
It is outside the installable skills.

Run it from the repository root after skills.sh has rescanned a change:

```sh
go -C maintenance/audits run .            # rewrite the table in README.md
go -C maintenance/audits run . -print     # print the table only
```

The default paths are relative to this directory; `-config` and `-readme`
override them.

Skill order comes from `skills.sh.json`. The table is written between the
`<!-- audits:start -->` and `<!-- audits:end -->` markers. A skill whose page
returns 404 is reported as not listed, and an audit with no result yet as
pending. Each result is a shields.io badge in the colour skills.sh uses for that status, linked to the audit's detail page. The tool reads the public skill pages; if their markup changes, the
parser in `main.go` needs to follow.
