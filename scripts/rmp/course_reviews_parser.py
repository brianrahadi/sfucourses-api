import os
import json
import glob
import re
from collections import defaultdict
from datetime import datetime

# Real SFU course codes (e.g. CMPT105W, SA150, FALX99), from the API's own outlines data.
OUTLINES_PATH = os.environ.get(
    "RMP_OUTLINES",
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "internal", "store", "json", "outlines.json"),
)
try:
    with open(OUTLINES_PATH, encoding="utf-8") as _f:
        SFU_COURSES = {(c["dept"] + c["number"]).upper() for c in json.load(_f)}
except Exception as _e:
    print(f"Warning: could not load {OUTLINES_PATH} ({_e}); using pattern check only")
    SFU_COURSES = set()

# 2-4 letter subject, 3 digits, optional suffix letter (W writing-intensive, etc.)
COURSE_CODE_PATTERN = re.compile(r'^[A-Z]{2,4}\d{3}[A-Z]?$')


def normalize_course_code(course_code):
    """'cmpt 105w', 'CMPT-105W' -> 'CMPT105W' (matches the API's normalizeCourseCode)."""
    return re.sub(r'[\s\-_./]', '', (course_code or '').upper())


def is_valid_course_code(course_code):
    """A real SFU course, or something shaped like one (older courses no longer in outlines).
    Examples: CMPT225, SA150, BUS360W, FALX99"""
    if not course_code:
        return False
    return course_code in SFU_COURSES or bool(COURSE_CODE_PATTERN.match(course_code))

def review_sort_key(review):
    """Newest first; unparseable dates sort last."""
    cleaned = re.sub(r'(\d+)(st|nd|rd|th)', r'\1', (review.get('date') or '').strip())
    try:
        return datetime.strptime(cleaned, '%b %d, %Y')
    except ValueError:
        return datetime.min

def parse_course_reviews():
    # Create course_reviews_json directory
    os.makedirs("course_reviews_json", exist_ok=True)
    
    # Find all JSON files in reviews_json folder structure
    reviews_json_dir = "reviews_json"
    if not os.path.exists(reviews_json_dir):
        print(f"Error: {reviews_json_dir} directory not found!")
        return
    
    # Get all JSON files recursively from reviews_json
    json_files = sorted(glob.glob(os.path.join(reviews_json_dir, "**", "*.json"), recursive=True))
    
    if not json_files:
        print(f"No JSON files found in {reviews_json_dir}")
        return
    
    print(f"Found {len(json_files)} JSON files to process")
    
    # Dictionary to store course data (no department separation)
    courses = defaultdict(lambda: {
        'course_code': '',
        'total_reviews': 0,
        'instructors': defaultdict(lambda: {
            'professor_id': '',
            'professor_name': '',
            'ratings': [],
            'difficulties': [],
            'review_count': 0,
            'reviews': []
        })
    })
    
    processed_count = 0
    
    # Process each JSON file
    for file_path in json_files:
        print(f"Processing: {file_path}")
        
        try:
            with open(file_path, "r", encoding="utf-8") as f:
                instructor_data = json.load(f)
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            continue
        
        professor_id = instructor_data.get('professor_id', '')
        professor_name = instructor_data.get('professor_name', '')
        
        if not professor_id:
            print(f"Warning: No professor_id found in {file_path}")
            continue
        
        # Process each review
        reviews_processed = 0
        for review in instructor_data.get('reviews', []):
            course_code = review.get('course_code', '')
            if not course_code:
                continue
            
            # Normalize course code to uppercase for consistent grouping
            course_code_upper = normalize_course_code(course_code)
            
            # Skip things that are not course codes (e.g. "100", "PHILXX1")
            if not is_valid_course_code(course_code_upper):
                print(f"  Skipping invalid course code: {course_code}")
                continue
            
            reviews_processed += 1
            
            # Convert rating and difficulty to float
            try:
                rating = float(review.get('rating', 0)) if review.get('rating') else 0.0
                difficulty = float(review.get('difficulty', 0)) if review.get('difficulty') else 0.0
            except (ValueError, TypeError):
                rating = 0.0
                difficulty = 0.0
            
            # Convert helpful votes to int
            try:
                helpful = int(review.get('helpful', 0)) if review.get('helpful') else 0
                not_helpful = int(review.get('not_helpful', 0)) if review.get('not_helpful') else 0
            except (ValueError, TypeError):
                helpful = 0
                not_helpful = 0
            
            # Create structured review
            structured_review = {
                "rating": rating,
                "difficulty": difficulty,
                "course_code": course_code_upper,
                "date": review.get('date', ''),
                "metadata": {
                    "For Credit": review.get('metadata', {}).get('For Credit', ''),
                    "Attendance": review.get('metadata', {}).get('Attendance', ''),
                    "Would Take Again": review.get('metadata', {}).get('Would Take Again', ''),
                    "Grade": review.get('metadata', {}).get('Grade', ''),
                    "Textbook": review.get('metadata', {}).get('Textbook', '')
                },
                "review_msg": review.get('review_msg', ''),
                "helpful": helpful,
                "not_helpful": not_helpful,
                "tags": review.get('tags', [])
            }
            
            # Add to course data (using uppercase course code as key)
            course_data = courses[course_code_upper]
            course_data['course_code'] = course_code_upper
            course_data['total_reviews'] += 1
            
            instructor_summary = course_data['instructors'][professor_id]
            
            # Only set professor info if not already set (avoid overwriting)
            if not instructor_summary['professor_id']:
                instructor_summary['professor_id'] = professor_id
            if not instructor_summary['professor_name']:
                instructor_summary['professor_name'] = professor_name
            
            instructor_summary['ratings'].append(rating)
            instructor_summary['difficulties'].append(difficulty)
            instructor_summary['review_count'] += 1
            instructor_summary['reviews'].append(structured_review)
        
        print(f"  Processed {reviews_processed} reviews for {professor_name} ({professor_id})")
        processed_count += 1
    
    # Save course data (no department separation)
    all_courses_data = []
    
    for course_code, course_data in courses.items():
        # Calculate averages for each instructor
        instructors_list = []
        course_all_ratings = []
        course_all_difficulties = []
        
        for prof_id, instructor_summary in course_data['instructors'].items():
            ratings = instructor_summary['ratings']
            difficulties = instructor_summary['difficulties']
            
            avg_rating = sum(ratings) / len(ratings) if ratings else 0.0
            avg_difficulty = sum(difficulties) / len(difficulties) if difficulties else 0.0
            
            # Collect all ratings and difficulties for course-level averages
            course_all_ratings.extend(ratings)
            course_all_difficulties.extend(difficulties)
            
            instructor_data = {
                "professor_id": instructor_summary['professor_id'],
                "professor_name": instructor_summary['professor_name'],
                "avg_rating": round(avg_rating, 2),
                "avg_difficulty": round(avg_difficulty, 2),
                "review_count": instructor_summary['review_count'],
                "reviews": sorted(instructor_summary['reviews'], key=review_sort_key, reverse=True)
            }
            instructors_list.append(instructor_data)
        
        # Calculate course-level averages
        course_avg_rating = sum(course_all_ratings) / len(course_all_ratings) if course_all_ratings else 0.0
        course_avg_difficulty = sum(course_all_difficulties) / len(course_all_difficulties) if course_all_difficulties else 0.0
        
        # Create aggregated course data for ALL_COURSES.json
        aggregated_course_data = {
            "course_code": course_data['course_code'],
            "total_reviews": course_data['total_reviews'],
            "avg_rating": round(course_avg_rating, 2),
            "avg_difficulty": round(course_avg_difficulty, 2)
        }
        all_courses_data.append(aggregated_course_data)
        
        # Create course review data
        course_review_data = {
            "course_code": course_data['course_code'],
            "total_reviews": course_data['total_reviews'],
            "instructors": instructors_list
        }
        
        # Save course data
        safe_course_code = "".join(c for c in course_code if c.isalnum() or c in (' ', '-', '_')).rstrip()
        safe_course_code = safe_course_code.replace(' ', '_')
        json_filename = f"{safe_course_code}.json"
        json_path = os.path.join("course_reviews_json", json_filename)
        
        try:
            with open(json_path, "w", encoding="utf-8") as json_file:
                json.dump(course_review_data, json_file, ensure_ascii=False)
            
            print(f"✅ Saved: {json_filename} ({course_data['total_reviews']} reviews, {len(instructors_list)} instructors)")
            
        except Exception as e:
            print(f"Error saving {json_filename}: {e}")
    
    # Sort ALL_COURSES.json by course_code
    all_courses_data.sort(key=lambda x: x['course_code'])
    
    # Save ALL_COURSES.json
    all_courses_path = os.path.join("course_reviews_json", "ALL_COURSES.json")
    try:
        with open(all_courses_path, "w", encoding="utf-8") as json_file:
            json.dump(all_courses_data, json_file, ensure_ascii=False)
        
        print(f"✅ Saved: ALL_COURSES.json ({len(all_courses_data)} courses)")
        
    except Exception as e:
        print(f"Error saving ALL_COURSES.json: {e}")
    
    print(f"\n🎉 Processing complete! Processed {processed_count} instructor files.")
    print(f"Course review files saved in course_reviews_json/ folder")

if __name__ == "__main__":
    parse_course_reviews()

