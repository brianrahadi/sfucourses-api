"""Stage 1: scrape SFU professor listings per RMP department.

Run from the work directory. Writes scraped_general/<Department>.csv,
scraped_general/ALL.csv and scraped_general/ALL.json.
Departments that are commented out keep the CSV that run.sh seeds from the
previous all_instructor_reviews.json.
"""
import os
import csv
import json
import time
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from driver import make_driver

BASE_URL = "https://www.ratemyprofessors.com/search/professors/1482?did="
OUTPUT_DIR = "scraped_general"
os.makedirs(OUTPUT_DIR, exist_ok=True)

links = [
    {"Accounting": 1},
    {"Advisor": 1136},
    {"Agriculture":  2},
    {"Anthropology": 3},
    {"Applied Arts": 496},
    {"Archaeology": 829},
    {"Art": 113},
    {"Art & Design": 308},
    {"Art Education": 203},
    {"Asian Studies": 177},
    {"Biochemistry": 93},
    {"Biological Sciences": 189},
    {"Biology": 6},
    {"Biomedical Physiology & Kinesiology": 2524},
    {"Business": 3},
    {"Business Administration": 154},
    {"Business Economics": 613},
    {"Canadian Studies": 345},
    {"Chemistry": 8},
    {"Cognitive Science": 1052},
    {"Communication": 10},
    {"Computer & Informational Tech.": 355},
    {"Computer Science": 11},
    {"Counseling": 163},
    {"Counseling & Educational Psych": 157},
    {"Criminal Justice": 12},
    {"Criminology": 311},
    {"Culinary Arts": 13},
    {"Design": 3},
    {"Earth Science": 78},
    {"Economics": 15},
    {"Education": 16},
    {"Educational Research": 590}, # 1 prof
    {"Engineering": 17},
    {"English": 18},
    # {"English & Liberal Studies": 904}, # 1 prof
    # {"Environmental Public Health": 317}, # 1 prof
    {"Environmental Resources": 885},
    {"Environmental Science": 108},
    {"Environmental Studies": 264},
    {"Film": 20},
    {"Finance": 21},
    {"Fine Arts": 22},
    {"Foreign Languages": 281},
    {"Forensic Science": 1389},
    {"French": 70},
    {"Gender Studies": 1076},
    {"General Studies": 200},
    {"Geography": 23},
    {"Geology": 24},
    {"German": 65},
    {"Gerontology": 94},
    {"Graphic Arts": 25},
    # {"Graphic Design": 423}, # 1 prof
    # {"Health & Human Performance": 117}, # 1 prof
    {"Health Administration": 618}, # 1 prof
    {"Health Science": 26},
    {"History": 27},
    {"History & Political Science": 283},
    {"Humanities": 29},
    {"Indigenous Studies": 2310},
    # {"Industrial Studies": 1083}, # 2 prof
    # {"Information Systems": 246}, # 1 prof
    {"Information Technology": 381},
    {"Interactive Media Design": 883},
    # {"Internation Service": 494}, # 1 prof
    {"International Business": 169},
    {"International Studies": 135},
    {"Italian": 81},
    {"Justice Studies": 1023},
    {"Kinesiology": 333},
    {"Labor Studies": 759},
    {"Languages": 33},
    {"Law": 34},
    # {"Learning Center": 102}, 1 prof
    {"Liberal Studies": 532},
    {"Linguistics": 129},
    {"Literature": 35},
    {"Management": 36},
    {"Management Information Systems": 2000},
    {"Marketing": 37},
    {"Mathematics": 38},
    # {"Media & Performing Arts": 430}, # 1
    # {"Media Arts": 749}, # 1 prof
    # {"Media Studies": 180}, # 1 prof
    {"Microbiology": 134}, 
    {"Molecular Biology & Biochemistry": 2684},
    {"Molecular Biosciences": 1268},
    {"Music": 40},
    {"Native Studies": 344},
    {"Philosophy": 42},
    {"Physical Education": 43},
    {"Physics": 44},
    {"Physiology": 772},
    {"Policy Studies": 835},
    {"Political Science": 45},
    {"Psychology": 46},
    {"Psychology & Sociology": 275},
    {"Public Policy": 631},
    {"Publishing & Printing": 221},
    {"Religion": 47},
    {"Religious Studies": 339},
    {"Resource Management": 1720},
    {"Science": 48},
    {"Social Science": 49},
    {"Sociology": 51},
    {"Sociology & Anthropology": 1846},
    {"Statistics": 140},
    # {"Ta": 533},
    {"Theater": 52},
    {"Urban Planning": 354},
    {"Women": 2845},
    {"Women's Studies": 53},
    {"Writing": 54}
]


def extract_card_data(card):
    data = {}
    href = card.get_attribute("href")
    data["URL"] = href if href else "N/A"

    def text_of(selector, idx=0):
        elems = card.find_elements(By.CSS_SELECTOR, selector)
        return elems[idx].text.strip() if len(elems) > idx else "N/A"

    data["Quality"] = text_of(".CardNumRating__CardNumRatingNumber-sc-17t4b9u-2")
    ratings = text_of(".CardNumRating__CardNumRatingCount-sc-17t4b9u-3")
    data["Ratings"] = ratings.split()[0] if ratings != "N/A" and ratings else "N/A"
    data["Name"] = text_of(".CardName__StyledCardName-sc-1gyrgim-0")
    data["WouldTakeAgain"] = text_of(".CardFeedback__CardFeedbackItem-lq6nix-1 .CardFeedback__CardFeedbackNumber-lq6nix-2")
    data["Difficulty"] = text_of(".CardFeedback__CardFeedbackNumber-lq6nix-2", 1)
    return data


def scrape_department(driver, dept_name, dept_id):
    wait = WebDriverWait(driver, 10)
    driver.get(BASE_URL + str(dept_id))
    time.sleep(2)

    while True:
        try:
            btn = wait.until(EC.element_to_be_clickable((By.XPATH, "//button[contains(text(),'Show More')]")))
            driver.execute_script("arguments[0].click();", btn)
            time.sleep(1)
        except Exception:
            break

    cards = driver.find_elements(By.CSS_SELECTOR, "a[href*='/professor/']")
    rows = []
    for card in cards:
        try:
            rows.append(extract_card_data(card))
        except Exception as e:
            print(f"  failed to read a card: {e}")
    print(f"{dept_name}: found {len(rows)} professors")

    if not rows:
        # Keep the previous CSV rather than wiping a department on a bad page load.
        print(f"  WARNING: no professors found for {dept_name}; keeping previous CSV")
        return

    outpath = os.path.join(OUTPUT_DIR, f"{dept_name}.csv")
    with open(outpath, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=["URL", "Quality", "Ratings", "Name", "WouldTakeAgain", "Difficulty"])
        writer.writeheader()
        writer.writerows(rows)


def combine():
    fieldnames = ["URL", "Quality", "Ratings", "Name", "WouldTakeAgain", "Difficulty", "Department"]
    all_data = []
    csv_files = sorted(f for f in os.listdir(OUTPUT_DIR) if f.endswith(".csv") and f != "ALL.csv")
    for filename in csv_files:
        dept_name = filename[:-4]
        with open(os.path.join(OUTPUT_DIR, filename), encoding="utf-8") as f:
            for row in csv.DictReader(f):
                row["Department"] = dept_name
                all_data.append(row)

    with open(os.path.join(OUTPUT_DIR, "ALL.csv"), "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(all_data)
    with open(os.path.join(OUTPUT_DIR, "ALL.json"), "w", encoding="utf-8") as f:
        json.dump(all_data, f, ensure_ascii=False)
    print(f"Combined {len(all_data)} records from {len(csv_files)} departments")


if __name__ == "__main__":
    driver = make_driver()
    try:
        for entry in links:
            dept_name, dept_id = list(entry.items())[0]
            try:
                scrape_department(driver, dept_name, dept_id)
            except Exception as e:
                print(f"{dept_name}: error {e}; keeping previous CSV")
    finally:
        driver.quit()
    combine()
