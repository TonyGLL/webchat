from playwright.sync_api import sync_playwright, expect

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    try:
        # Login
        page.goto("http://localhost:3000/auth/login")
        page.wait_for_selector('label:has-text("Email")', timeout=60000)
        page.get_by_label("Email").fill("test@example.com")
        page.get_by_label("Password").fill("password")
        page.get_by_role("button", name="Login").click()
        expect(page).to_have_url("http://localhost:3000/dashboard", timeout=10000)

        # Verify Dashboard
        expect(page.get_by_text("Rooms")).to_be_visible()
        page.screenshot(path="jules-scratch/verification/dashboard.png")

        # Verify Profile
        page.goto("http://localhost:3000/profile")
        expect(page.get_by_text("User Profile")).to_be_visible()
        page.screenshot(path="jules-scratch/verification/profile.png")

    finally:
        browser.close()

with sync_playwright() as playwright:
    run(playwright)