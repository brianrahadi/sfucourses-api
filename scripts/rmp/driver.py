"""Shared headless Chrome setup for CI (GitHub Actions ubuntu runners ship Chrome)."""
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

USER_AGENT = (
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
)


def make_driver():
    options = Options()
    options.add_argument("--headless=new")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--window-size=1400,1000")
    options.add_argument(f"--user-agent={USER_AGENT}")
    # Speed: don't wait for ads/trackers to finish loading, and skip images.
    options.page_load_strategy = "eager"
    options.add_argument("--blink-settings=imagesEnabled=false")
    options.add_experimental_option("prefs", {"profile.managed_default_content_settings.images": 2})
    # Selenium Manager (bundled with selenium>=4.6) resolves the matching chromedriver.
    driver = webdriver.Chrome(options=options)
    driver.set_page_load_timeout(60)
    return driver
