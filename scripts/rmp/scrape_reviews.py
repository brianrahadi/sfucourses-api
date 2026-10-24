"""Stage 2: save RMP professor pages (with ratings expanded) as HTML.

Run from the work directory after scrape_listings.py.

Incremental: a professor is only fetched when the rating count on the listing
differs from the total_ratings saved in reviews_json/ (seeded from the repo).
For professors we already have, "Load More Ratings" stops once it reaches
reviews older than (newest saved review - OVERLAP_DAYS); reviews_parser.py then
merges the new page with the saved JSON, so older reviews are kept.

Writes reviews_general/<Department>/<prof_id>.html
"""
import csv
import glob
import json
import os
import random
import re
import shutil
import threading
import time
from datetime import datetime, timedelta

from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait

from driver import make_driver

# Earliest reviews to load for professors with no saved data
BASE_CUTOFF = datetime(2024, 1, 1)
# Re-read this many days before the newest saved review, to be safe
OVERLAP_DAYS = 45
THREADS = int(os.environ.get("RMP_THREADS", "4"))
# Set RMP_FULL=1 to ignore saved data and re-fetch everyone back to BASE_CUTOFF
FULL = os.environ.get("RMP_FULL") == "1"

BUTTON_XPATH = "//button[text()='Load More Ratings']"
TIMESTAMP_CSS = "div[class*='TimeStamp__StyledTimeStamp']"
NO_RATINGS_CSS = "div[class*='NoRatingsArea__StyledNoRatingsArea']"

done_lock = threading.Lock()
done_pages = {}  # prof_id -> html path already saved (same prof listed in several departments)
stats = {"saved": 0, "skipped_old": 0, "no_ratings": 0, "errors": 0}


def parse_rmp_date(date_str):
    cleaned = re.sub(r"(\d+)(st|nd|rd|th)", r"\1", date_str.strip())
    return datetime.strptime(cleaned, "%b %d, %Y")


def load_saved():
    """prof_id -> (set of saved total_ratings, newest saved review date or None)"""
    saved = {}
    for path in glob.glob(os.path.join("reviews_json", "*", "*.json")):
        try:
            with open(path, encoding="utf-8") as f:
                data = json.load(f)
        except Exception:
            continue
        pid = str(data.get("professor_id", ""))
        if not pid:
            continue
        totals, newest = saved.get(pid, (set(), None))
        totals.add(str(data.get("total_ratings")))
        for r in data.get("reviews", []):
            try:
                d = parse_rmp_date(r.get("date") or "")
            except Exception:
                continue
            if newest is None or d > newest:
                newest = d
        saved[pid] = (totals, newest)
    return saved


def build_jobs(saved):
    jobs = []  # (department, url, prof_id, cutoff)
    for csv_path in sorted(glob.glob(os.path.join("scraped_general", "*.csv"))):
        dept = os.path.basename(csv_path)[:-4]
        if dept == "ALL":
            continue
        with open(csv_path, encoding="utf-8") as f:
            for row in csv.DictReader(f):
                url = row.get("URL", "N/A")
                ratings = (row.get("Ratings") or "").strip()
                if url == "N/A" or ratings in ("0", "", "N/A"):
                    continue
                pid = url.rstrip("/").split("/")[-1]
                totals, newest = saved.get(pid, (set(), None))
                if not FULL and ratings in totals:
                    continue
                cutoff = BASE_CUTOFF
                if not FULL and newest is not None:
                    cutoff = max(BASE_CUTOFF, newest - timedelta(days=OVERLAP_DAYS))
                jobs.append((dept, url, pid, cutoff))
    return jobs


def scrape_one(driver, dept, url, pid, cutoff):
    dept_folder = os.path.join("reviews_general", dept)
    os.makedirs(dept_folder, exist_ok=True)
    html_path = os.path.join(dept_folder, f"{pid}.html")

    with done_lock:
        existing = done_pages.get(pid)
    if existing:
        shutil.copyfile(existing, html_path)
        return

    driver.get(url)
    try:
        WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.CSS_SELECTOR, f"{NO_RATINGS_CSS}, {TIMESTAMP_CSS}"))
        )
    except Exception:
        pass

    if driver.find_elements(By.CSS_SELECTOR, NO_RATINGS_CSS):
        stats["no_ratings"] += 1
        return

    stamps = driver.find_elements(By.CSS_SELECTOR, TIMESTAMP_CSS)
    if stamps:
        try:
            if parse_rmp_date(stamps[0].text) < cutoff:
                stats["skipped_old"] += 1
                return
        except Exception:
            pass

    while True:
        try:
            driver.execute_script(
                "document.querySelectorAll('iframe, .il-sr-info-native-copy-grey')"
                ".forEach(function(a){a.style.display='none';});"
            )
            btn = WebDriverWait(driver, 5).until(EC.presence_of_element_located((By.XPATH, BUTTON_XPATH)))
            driver.execute_script("arguments[0].scrollIntoView({block: 'center'});", btn)
            driver.execute_script("arguments[0].click();", btn)
            time.sleep(random.uniform(1.0, 2.5))
        except Exception:
            break  # no more button
        stamps = driver.find_elements(By.CSS_SELECTOR, TIMESTAMP_CSS)
        try:
            if stamps and parse_rmp_date(stamps[-1].text) < cutoff:
                break
        except Exception:
            pass

    with open(html_path, "w", encoding="utf-8") as f:
        f.write(driver.page_source)
    with done_lock:
        done_pages[pid] = html_path
    stats["saved"] += 1


def worker(thread_id, jobs):
    driver = make_driver()
    try:
        for i, (dept, url, pid, cutoff) in enumerate(jobs, 1):
            try:
                scrape_one(driver, dept, url, pid, cutoff)
            except Exception as e:
                stats["errors"] += 1
                print(f"[t{thread_id}] error on {url}: {e}")
                # A crashed tab can poison the session; start a fresh browser.
                try:
                    driver.quit()
                except Exception:
                    pass
                driver = make_driver()
            if i % 25 == 0:
                print(f"[t{thread_id}] {i}/{len(jobs)}")
    finally:
        driver.quit()


def main():
    os.makedirs("reviews_general", exist_ok=True)
    saved = load_saved()
    jobs = build_jobs(saved)
    print(f"{len(jobs)} professor pages to fetch (full={FULL})")
    if not jobs:
        return
    # Keep the same professor in one thread so duplicates are copied, not refetched.
    jobs.sort(key=lambda j: j[2])
    buckets = [[] for _ in range(THREADS)]
    for job in jobs:
        buckets[int(job[2]) % THREADS if job[2].isdigit() else 0].append(job)
    threads = [threading.Thread(target=worker, args=(i + 1, b)) for i, b in enumerate(buckets) if b]
    for t in threads:
        t.start()
    for t in threads:
        t.join()
    print(f"Stage 2 done: {stats}")


if __name__ == "__main__":
    main()
