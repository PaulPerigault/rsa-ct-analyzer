# testdata/ct

RFC 6962 responses recorded from a public Certificate Transparency log. Public data: the entries
are certificates of cloud-provider infrastructure host names, chosen so that no person is
identifiable. The tests replay these files from a local `httptest` server; they never reach the
public network.

| Item | Value |
| --- | --- |
| Log | `https://ct.googleapis.com/logs/us1/argon2026h2` |
| Recorded | 2026-11-12, one `get-sth` and one `get-entries?start=1000000&end=1000031` (read-only `GET`) |
| `get-sth.json` | the Signed Tree Head as returned (tree size 3501428149, signature not verified by the tool) |
| `get-entries.json` | 4 of the 32 entries returned, kept in order: log indices 1000002 and 1000010 (`precert_entry`), 1000003 and 1000004 (`x509_entry`), all RSA-2048 |

The file was trimmed, not edited: each entry is byte-identical to the log's answer (`leaf_input`
and `extra_data`, base64). Entries are renumbered 0 to 3 in the file; a replaying server answers any
range with these four.

Entry structure (RFC 6962 §3.4 and §4.6):

- `x509_entry`: the leaf holds the certificate; `extra_data` holds the issuer chain.
- `precert_entry`: the leaf holds the issuer key hash and the TBSCertificate; `extra_data` holds the
  pre-certificate and its chain. The tool reads the key from the pre-certificate.

To re-record, request the same two URLs once, trim, and update the table above.
