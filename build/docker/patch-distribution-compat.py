#!/usr/bin/env python3
"""Adapt two legacy distribution dependencies without changing storage features."""

from pathlib import Path
import sys


LEGACY = '''\tfor _, v := range acls.Items {
\t\tif m, ok := v.(map[string]interface{}); ok {
\t\t\tentity, ok1 := m["entity"].(string)
\t\t\trole, ok2 := m["role"].(string)
\t\t\tif ok1 && ok2 {
\t\t\t\tr = append(r, ACLRule{Entity: ACLEntity(entity), Role: ACLRole(role)})
\t\t\t}
\t\t}
\t}'''

TYPED = '''\tfor _, v := range acls.Items {
\t\tif v != nil && v.Entity != "" && v.Role != "" {
\t\t\tr = append(r, ACLRule{Entity: ACLEntity(v.Entity), Role: ACLRole(v.Role)})
\t\t}
\t}'''


def patch(root: Path) -> None:
    changes = []
    for relative, legacy, modern, count in (
        ("vendor/google.golang.org/cloud/storage/acl.go", LEGACY, TYPED, 2),
        ("vendor/github.com/xenolf/lego/acme/dns_challenge.go",
         "if err == dns.ErrTruncated {", "if in != nil && in.Truncated {", 1),
    ):
        source = root / relative
        text = source.read_text()
        if text.count(legacy) == 0 and text.count(modern) == count:
            continue
        if text.count(legacy) != count or text.count(modern) != 0:
            raise ValueError(f"unexpected legacy source in {relative}; refusing partial patch")
        changes.append((source, text.replace(legacy, modern)))
    # Check both source contracts before modifying either file.
    for source, text in changes:
        source.write_text(text)


if __name__ == "__main__":
    patch(Path(sys.argv[1]) if len(sys.argv) > 1 else Path.cwd())
