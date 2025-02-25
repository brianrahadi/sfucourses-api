.PHONY: fetch-outlines
fetch-outlines:
	go run scripts/fetchOutlines/main.go

.PHONY: fetch-sections
fetch-sections:
	go run scripts/fetchSections/main.go

.PHONY: sync-offerings
sync-offerings:
	go run scripts/syncOfferings/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt