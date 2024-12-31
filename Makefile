.PHONY: fetch-outlines
fetch-outlines:
	go run scripts/fetchOutlines/main.go

.PHONY: fetch-courses
fetch-courses:
	go run scripts/fetchCourses/main.go

.PHONY: sync-offerings
sync-offerings:
	go run scripts/syncOfferings/main.go