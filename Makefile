.PHONY: fetch-outlines
fetch-outlines:
	go run scripts/fetchOutlines/main.go

.PHONY: fetch-sections
fetch-sections:
	go run scripts/fetchSections/main.go