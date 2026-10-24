#!/usr/bin/env bash
# Refresh RateMyProfessors review data in internal/store/json.
#
#   scripts/rmp/run.sh            # incremental (default)
#   RMP_FULL=1 scripts/rmp/run.sh # re-fetch every professor back to 2024
#
# Pipeline (all in a temp work dir, seeded from the repo's current data):
#   1. scrape_listings.py      -> scraped_general/*.csv, ALL.json
#   2. scrape_reviews.py       -> reviews_general/<dept>/<id>.html (changed profs only)
#   3. reviews_parser.py       -> reviews_json/ (merged with existing reviews)
#   4. course_reviews_parser.py-> course_reviews_json/, ALL_COURSES.json
#   then copy results back into internal/store/json.
set -euo pipefail

SCRIPTS="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$SCRIPTS/../.." && pwd)"
DATA="$REPO/internal/store/json"
WORK="${RMP_WORKDIR:-$(mktemp -d)}"
PY="${PYTHON:-python3}"

echo "== Work dir: $WORK"
mkdir -p "$WORK"
cd "$WORK"

echo "== Seeding from repo data"
rm -rf reviews_json scraped_general reviews_general course_reviews_json
cp -R "$DATA/instructor_reviews" reviews_json
"$PY" - "$DATA/all_instructor_reviews.json" <<'EOF'
import csv, json, os, sys
from collections import defaultdict
rows = json.load(open(sys.argv[1], encoding="utf-8"))
by_dept = defaultdict(list)
for r in rows:
    by_dept[r["Department"]].append(r)
os.makedirs("scraped_general", exist_ok=True)
fields = ["URL", "Quality", "Ratings", "Name", "WouldTakeAgain", "Difficulty"]
for dept, items in by_dept.items():
    with open(os.path.join("scraped_general", f"{dept}.csv"), "w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=fields, extrasaction="ignore")
        w.writeheader()
        w.writerows(items)
print(f"Seeded {len(rows)} professors in {len(by_dept)} departments")
EOF

count() { "$PY" -c "import json,sys; print(len(json.load(open(sys.argv[1]))))" "$1"; }
PREV_PROFS=$(count "$DATA/all_instructor_reviews.json")
PREV_COURSES=$(count "$DATA/all_course_reviews.json")

echo "== Stage 1: department listings"
"$PY" "$SCRIPTS/scrape_listings.py"
NEW_PROFS=$(count scraped_general/ALL.json)
echo "Professors: $PREV_PROFS -> $NEW_PROFS"
if [ "$NEW_PROFS" -lt $((PREV_PROFS * 9 / 10)) ]; then
  echo "::error::Stage 1 found far fewer professors than before (blocked or RMP layout changed). Aborting."
  exit 1
fi

echo "== Stage 2: professor pages"
"$PY" "$SCRIPTS/scrape_reviews.py"

echo "== Stage 3: parse reviews"
"$PY" "$SCRIPTS/reviews_parser.py" > stage3.log
tail -n 3 stage3.log

echo "== Stage 4: course reviews"
"$PY" "$SCRIPTS/course_reviews_parser.py" > stage4.log
tail -n 3 stage4.log
NEW_COURSES=$(count course_reviews_json/ALL_COURSES.json)
echo "Courses: $PREV_COURSES -> $NEW_COURSES"
if [ "$NEW_COURSES" -lt $((PREV_COURSES * 95 / 100)) ]; then
  echo "::error::Course count dropped sharply. Aborting without touching repo data."
  exit 1
fi

echo "== Writing results into internal/store/json"
rm -rf "$DATA/course_reviews" "$DATA/instructor_reviews"
cp -R course_reviews_json "$DATA/course_reviews"
cp -R reviews_json "$DATA/instructor_reviews"
cp course_reviews_json/ALL_COURSES.json "$DATA/all_course_reviews.json"
cp scraped_general/ALL.json "$DATA/all_instructor_reviews.json"
echo "== Done"
