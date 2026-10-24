import os
import glob
import json
from bs4 import BeautifulSoup

def parse_reviews():
    # Create reviews_json directory
    os.makedirs("reviews_json", exist_ok=True)
    
    # Find all HTML files in reviews_general folder structure
    reviews_general_dir = "reviews_general"
    if not os.path.exists(reviews_general_dir):
        print(f"Error: {reviews_general_dir} directory not found!")
        return
    
    # Get all HTML files recursively from reviews_general
    html_files = sorted(glob.glob(os.path.join(reviews_general_dir, "**", "*.html"), recursive=True))
    
    if not html_files:
        print(f"No HTML files found in {reviews_general_dir}")
        return
    
    print(f"Found {len(html_files)} HTML files to process")
    
    processed_count = 0
    
    # Process each HTML file
    for file_path in html_files:
        print(f"Processing: {file_path}")
        
        try:
            with open(file_path, "r", encoding="utf-8") as file:
                soup = BeautifulSoup(file, "html.parser")
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            continue
        
        # Extract professor ID from filename (e.g., "12345.html" → "12345")
        professor_id = os.path.splitext(os.path.basename(file_path))[0]
        
        # Extract department from folder path
        path_parts = file_path.split(os.sep)
        department = path_parts[-2] if len(path_parts) >= 2 else "Unknown"
        
        # Extract professor-level stats
        professor_name_elem = soup.find("h1", class_="NameTitle__NameWrapper-dowf0z-2 cSXRap")
        overall_rating_elem = soup.find("div", class_="RatingValue__Numerator-qw8sqy-2 duhvlP")
        
        feedback_nums_elem = soup.find("div", class_="TeacherFeedback__StyledTeacherFeedback-gzhlj7-0 fnfHKd")
        if feedback_nums_elem is None:
            print(f"No profile data found for {professor_id}, skipping...")
            continue
        
        feedback_numbers = feedback_nums_elem.find_all("div", class_="FeedbackItem__FeedbackNumber-uof32n-1 ecFgca")
        
        would_take_again_percentage = feedback_numbers[0].text.strip().replace("%", "") if len(feedback_numbers) > 0 else None
        difficulty_level = feedback_numbers[1].text.strip() if len(feedback_numbers) > 1 else None
        
        department_elem = soup.find("a", class_="TeacherDepartment__StyledDepartmentLink-fl79e8-0 fChJtl")
        total_ratings_elem = soup.find("a", href="#ratingsList")
        
        # Initialize professor data structure
        professor_data = {
            "professor_id": professor_id,
            "professor_name": professor_name_elem.text.strip() if professor_name_elem else None,
            "overall_rating": overall_rating_elem.text.strip() if overall_rating_elem else None,
            "would_take_again": would_take_again_percentage,
            "difficulty_level": difficulty_level,
            "department": department_elem.text.strip() if department_elem else department,
            "total_ratings": total_ratings_elem.text.strip().split()[0] if total_ratings_elem else None,
            "reviews": []
        }
        
        # Find the main ratings list
        ratings_list = soup.find(id="ratingsList")
        
        if ratings_list is None:
            print(f"No reviews found for {professor_id}")
            # Still save the professor data even if no reviews
        else:
            # Loop through all reviews inside ratingsList
            for review in ratings_list.find_all("li", recursive=False):
                # Skip ad elements
                if review.find("div") and "AdNoBid" in review.find("div").get('class', [''])[0]:
                    continue
                
                # Extract rating and difficulty
                rating_elements = review.find_all("div", class_="CardNumRating__CardNumRatingNumber-sc-17t4b9u-2")
                
                rating_value = rating_elements[0] if len(rating_elements) > 0 else None
                difficulty_value = rating_elements[1] if len(rating_elements) > 1 else None
                
                course_code_elem = review.find("div", class_="RatingHeader__StyledClass-sc-1dlkqw1-3")
                date_elem = review.find("div", class_="TimeStamp__StyledTimeStamp-sc-9q2r30-0")
                
                # Extract metadata (attendance, grade, etc.)
                metadata = {}
                metadata_blocks = review.find_all("div", class_="MetaItem__StyledMetaItem-y0ixml-0")
                
                for meta in metadata_blocks:
                    parts = meta.text.split(":")
                    if len(parts) == 2:
                        key = parts[0].strip()
                        value = parts[1].strip()
                        metadata[key] = value
                
                # Extract review message
                review_msg_elem = review.find("div", class_="Comments__StyledComments-dzzyvm-0")
                
                # Extract helpful/unhelpful votes
                helpful_votes = review.find_all("div", class_="Thumbs__HelpTotalNumber-sc-19shlav-2")
                helpful = helpful_votes[0].text.strip() if len(helpful_votes) > 0 else "0"
                not_helpful = helpful_votes[1].text.strip() if len(helpful_votes) > 1 else "0"
                
                # Extract review tags
                tags_elem = review.find("div", class_="RatingTags__StyledTags-sc-1boeqx2-0")
                tag_list = [tag.text for tag in tags_elem.find_all("span")] if tags_elem else []
                
                # Store review in professor's reviews list
                professor_data["reviews"].append({
                    "rating": rating_value.text.strip() if rating_value else None,
                    "difficulty": difficulty_value.text.strip() if difficulty_value else None,
                    "course_code": course_code_elem.text.strip() if course_code_elem else None,
                    "date": date_elem.text.strip() if date_elem else None,
                    "metadata": metadata,
                    "review_msg": review_msg_elem.text.strip() if review_msg_elem else None,
                    "helpful": helpful,
                    "not_helpful": not_helpful,
                    "tags": tag_list
                })
        
        # Create department folder
        dept_folder = os.path.join("reviews_json", department)
        os.makedirs(dept_folder, exist_ok=True)
        
        # Create safe filename from professor name
        professor_name = professor_data["professor_name"]
        if professor_name:
            # Replace invalid filename characters
            safe_filename = "".join(c for c in professor_name if c.isalnum() or c in (' ', '-', '_')).rstrip()
            safe_filename = safe_filename.replace(' ', '_')
            json_filename = f"{safe_filename}.json"
        else:
            json_filename = f"professor_{professor_id}.json"
        
        # Save professor data as JSON in department folder
        json_path = os.path.join(dept_folder, json_filename)
        
        # Merge with existing JSON to preserve old reviews (from before scraper cutoff)
        if os.path.exists(json_path):
            try:
                with open(json_path, "r", encoding="utf-8") as existing_file:
                    existing_data = json.load(existing_file)
                
                # Deduplicate newly parsed reviews first
                seen_keys = set()
                unique_reviews = []
                for r in professor_data["reviews"]:
                    key = (r.get("date", ""), r.get("course_code", ""), r.get("review_msg", ""))
                    if key not in seen_keys:
                        seen_keys.add(key)
                        unique_reviews.append(r)
                professor_data["reviews"] = unique_reviews
                
                # Append old reviews that aren't in the new set
                old_reviews_added = 0
                for old_review in existing_data.get("reviews", []):
                    key = (old_review.get("date", ""), old_review.get("course_code", ""), old_review.get("review_msg", ""))
                    if key not in seen_keys:
                        seen_keys.add(key)
                        professor_data["reviews"].append(old_review)
                        old_reviews_added += 1
                
                if old_reviews_added > 0:
                    print(f"  Merged {old_reviews_added} old reviews from existing JSON")
                    
            except Exception as e:
                print(f"  Warning: Could not merge existing JSON for {professor_id}: {e}")
        
        try:
            with open(json_path, "w", encoding="utf-8") as json_file:
                json.dump(professor_data, json_file, ensure_ascii=False)
            
            print(f"✅ Saved: {json_filename} ({len(professor_data['reviews'])} reviews)")
            processed_count += 1
            
        except Exception as e:
            print(f"Error saving {json_filename}: {e}")
    
    print(f"\n🎉 Processing complete! Processed {processed_count} professors.")
    print(f"JSON files saved in reviews_json/ folder")

if __name__ == "__main__":
    parse_reviews()
