# README audit table

`audits` rewrites the audit table in the repository `README.md` from the
results [skills.sh](https://skills.sh/sproutwerk/skills) shows for each skill.
It is outside the installable skills.

Run it from the repository root:

```sh
go -C maintenance/audits run . check      # has skills.sh rescanned every skill?
go -C maintenance/audits run .            # rewrite the table in README.md
go -C maintenance/audits run . -print     # print the table only
```

`check` compares the scan time on each audit's detail page with the newest
commit under that skill's directory. It lists every audit that is older than
its skill, not listed, or still without a result, and exits non-zero until none
is left. Scan times are read as UTC and are shown to the minute.

The default paths are relative to this directory; `-config` and `-readme`
override them.

Skill order comes from `skills.sh.json`. The table is written between the
`<!-- audits:start -->` and `<!-- audits:end -->` markers. A skill whose page
returns 404 is reported as not listed, and an audit with no result yet as
pending. Each result is a shields.io badge in the colour skills.sh uses for that status, linked to the audit's detail page. The tool reads the public skill pages; if their markup changes, the
parser in `main.go` needs to follow.
