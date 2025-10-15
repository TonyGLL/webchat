from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()
    page.goto("http://localhost:3000/register")
    page.get_by_label("Username").fill("testuser")
    page.get_by_label("Email Address").fill("test@example.com")
    page.get_by_label("Password").fill("password")
    page.get_by_role("button", name="Sign Up").click()
    page.wait_for_url("http://localhost:3000/verify-email?email=test@example.com")
    browser.close()

with sync_playwright() as playwright:
    run(playwright)