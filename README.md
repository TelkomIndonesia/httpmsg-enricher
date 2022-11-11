# HTTP Message Enricher

A simple service that can parse recorded HTTP request and response and provide metadata enrichment. Current enrichment includes: OWASP ModSecurity Core Rule Set, GeoIP, mime, and user-agent. Sample record available in [testdata files](./testdata/record.txt) and currently only [two API operations](./main.go#L40-L88) available.
