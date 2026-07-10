// ── Configuration ─────────────────────────────────────────────
// Replace this URL after you deploy your Lambda + API Gateway.
// Example: "https://abc123.execute-api.us-east-1.amazonaws.com/check"
const API_URL = "https://YOUR_API_GATEWAY_URL/check";

// ── Main function: called when user clicks "Run Checks" ────────
async function runChecks() {
  const prUrl   = document.getElementById("pr-url-input").value.trim();
  const token   = document.getElementById("token-input").value.trim();
  const btn     = document.getElementById("run-checks-btn");
  const btnText = document.getElementById("btn-text");
  const spinner = document.getElementById("btn-spinner");
  const errorBox     = document.getElementById("error-box");
  const resultsSection = document.getElementById("results-section");

  // Basic validation
  if (!prUrl) {
    showError("Please enter a GitHub PR URL.");
    return;
  }
  if (!prUrl.match(/github\.com\/.+\/.+\/pull\/\d+/)) {
    showError("URL doesn't look right. Use: https://github.com/owner/repo/pull/123");
    return;
  }

  // Reset UI
  hideError();
  resultsSection.classList.add("hidden");
  btn.disabled = true;
  btnText.classList.add("hidden");
  spinner.classList.remove("hidden");

  try {
    const response = await fetch(API_URL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ pr_url: prUrl, token: token }),
    });

    const data = await response.json();

    if (data.error) {
      showError(data.error);
      return;
    }

    renderResults(data.results);
  } catch (err) {
    showError(
      "Could not reach the PR Guardian API. " +
      "Make sure the API Gateway URL is configured in app.js, " +
      "or check your network connection.\n\nError: " + err.message
    );
  } finally {
    btn.disabled = false;
    btnText.classList.remove("hidden");
    spinner.classList.add("hidden");
  }
}

// ── Render check results ───────────────────────────────────────
function renderResults(results) {
  const section  = document.getElementById("results-section");
  const grid     = document.getElementById("results-grid");
  const summary  = document.getElementById("results-summary");
  const progress = document.getElementById("progress-bar");

  const passed = results.filter(r => r.passed).length;
  const total  = results.length;
  const pct    = Math.round((passed / total) * 100);

  // Summary text + colour
  let summaryColor = "#f85149";
  if (passed === total) summaryColor = "#3fb950";
  else if (passed >= total / 2) summaryColor = "#d29922";

  summary.innerHTML = `<span style="color:${summaryColor}">●</span> ${passed} / ${total} checks passed`;
  progress.style.width = pct + "%";
  progress.style.background = passed === total
    ? "linear-gradient(90deg, #238636, #3fb950)"
    : passed >= total / 2
      ? "linear-gradient(90deg, #9e6a03, #d29922)"
      : "linear-gradient(90deg, #8b0000, #f85149)";

  // Clear old results
  grid.innerHTML = "";

  // Render each check row with a slight stagger
  results.forEach((r, i) => {
    const row = document.createElement("div");
    row.className = `result-row ${r.passed ? "pass" : "fail"}`;
    row.style.animationDelay = `${i * 60}ms`;

    const icon = r.passed ? "✅" : "❌";
    row.innerHTML = `
      <div class="result-icon">${icon}</div>
      <div>
        <div class="result-name">${escapeHtml(r.name)}</div>
        <div class="result-msg">${escapeHtml(r.message)}</div>
      </div>
    `;
    grid.appendChild(row);
  });

  section.classList.remove("hidden");
  section.scrollIntoView({ behavior: "smooth", block: "nearest" });
}

// ── Error helpers ──────────────────────────────────────────────
function showError(msg) {
  const box = document.getElementById("error-box");
  box.textContent = msg;
  box.classList.remove("hidden");
}

function hideError() {
  document.getElementById("error-box").classList.add("hidden");
}

// ── Copy code snippet to clipboard ────────────────────────────
function copyCode(elementId, btnId) {
  const code = document.getElementById(elementId).textContent;
  navigator.clipboard.writeText(code).then(() => {
    const btn = document.getElementById(btnId);
    btn.textContent = "Copied!";
    btn.classList.add("copied");
    setTimeout(() => {
      btn.textContent = "Copy";
      btn.classList.remove("copied");
    }, 2000);
  });
}

// ── Utility: escape HTML to prevent XSS ───────────────────────
function escapeHtml(str) {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

// ── Allow pressing Enter in input to trigger check ────────────
document.addEventListener("DOMContentLoaded", () => {
  ["pr-url-input", "token-input"].forEach(id => {
    document.getElementById(id).addEventListener("keydown", e => {
      if (e.key === "Enter") runChecks();
    });
  });
});
