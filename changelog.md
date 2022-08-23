# Changelog

# v1.0.4 (2022-08-23)

- Use plain Action array rather than array of pooled pointers (#13)
- Introduce generic logger interface (#12)
- Update to Go 1.19 (#14)

# v1.0.3 (2021-12-16)

- support for parsing IPv6 addresses (#11)

# v1.0.2 (2021-06-01)

- fix-oom-on-http-requests (#9)

# v1.0.1 (2021-01-07)

- add license
- bugfixes
- fix random error under load 
- move buffer out of loop to reduce mem usage
- by default frame don't have actions and doesn't have ownership on it

# v1.0.0

- initial version
    - fix bug with encode/decode binary data