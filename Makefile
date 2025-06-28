.PHONY: build-all
build-all: build-fetch-sections build-fetch-outlines build-sync-offerings build-sync-instructors

.PHONY: build-fetch-sections
build-fetch-sections:
	go build -o bin/fetch-sections scripts/fetchSections/main.go

.PHONY: build-fetch-outlines
build-fetch-outlines:
	go build -o bin/fetch-outlines scripts/fetchOutlines/main.go

.PHONY: build-sync-offerings
build-sync-offerings:
	go build -o bin/sync-offerings scripts/syncOfferings/main.go

.PHONY: build-sync-instructors
build-sync-instructors:
	go build -o bin/sync-instructors scripts/syncInstructors/main.go

.PHONY: build-fetch-instructors
build-fetch-instructors:
	go build -o bin/fetch-instructors scripts/fetchInstructors/main.go

.PHONY: fetch-outlines
fetch-outlines:
	go run scripts/fetchOutlines/main.go

.PHONY: fetch-sections
fetch-sections:
	go run scripts/fetchSections/main.go

.PHONY: sync-offerings
sync-offerings:
	go run scripts/syncOfferings/main.go

.PHONY: sync-instructors
sync-instructors:
	go run scripts/syncInstructors/main.go

.PHONY: fetch-instructors
fetch-instructors:
	go run scripts/fetchInstructors/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt